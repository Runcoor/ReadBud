package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain/draft"
	taskDomain "readbud/internal/domain/task"
	"readbud/internal/pipeline"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"
	pipelinePkg "readbud/internal/service/pipeline"
)

// Server wraps the Asynq server and registers pipeline handlers.
type Server struct {
	srv         *asynq.Server
	mux         *asynq.ServeMux
	client      *asynq.Client
	taskSvc     *service.TaskService
	draftRepo   postgres.ArticleDraftRepository
	blockRepo   postgres.ArticleBlockRepository
	sourceRepo  postgres.SourceDocumentRepository
	llmProvider adapter.LLMProvider
	imageSearch adapter.ImageSearchProvider
	imageGen    adapter.ImageGenProvider
	logger      *zap.Logger
}

// ServerConfig holds configuration for the Asynq worker server.
type ServerConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
}

// NewServer creates a new Asynq worker server.
func NewServer(
	cfg ServerConfig,
	taskSvc *service.TaskService,
	draftRepo postgres.ArticleDraftRepository,
	blockRepo postgres.ArticleBlockRepository,
	sourceRepo postgres.SourceDocumentRepository,
	llmProvider adapter.LLMProvider,
	imageSearch adapter.ImageSearchProvider,
	imageGen adapter.ImageGenProvider,
	logger *zap.Logger,
) *Server {
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Printf("[worker] error processing task %s: %v", task.Type(), err)
		}),
	})

	client := asynq.NewClient(redisOpt)

	return &Server{
		srv:         srv,
		mux:         asynq.NewServeMux(),
		client:      client,
		taskSvc:     taskSvc,
		draftRepo:   draftRepo,
		blockRepo:   blockRepo,
		sourceRepo:  sourceRepo,
		llmProvider: llmProvider,
		imageSearch: imageSearch,
		imageGen:    imageGen,
		logger:      logger,
	}
}

// RegisterHandlers registers all pipeline stage handlers.
func (s *Server) RegisterHandlers() {
	s.mux.HandleFunc(pipeline.TypeKeywordExpand, s.handleKeywordExpand)
	s.mux.HandleFunc(pipeline.TypeSourceSearch, s.handleSourceSearch)
	s.mux.HandleFunc(pipeline.TypeContentCrawl, s.handleContentCrawl)
	s.mux.HandleFunc(pipeline.TypeHotScore, s.handleHotScore)
	s.mux.HandleFunc(pipeline.TypeArticleWrite, s.handleArticleWrite)
	s.mux.HandleFunc(pipeline.TypeImageMatch, s.handleImageMatch)
	s.mux.HandleFunc(pipeline.TypeChartGen, s.handleChartGen)
	s.mux.HandleFunc(pipeline.TypeHTMLCompile, s.handleHTMLCompile)
	s.mux.HandleFunc(pipeline.TypePublish, s.handlePublish)
}

// Start starts the Asynq server.
func (s *Server) Start() error {
	s.RegisterHandlers()
	return s.srv.Start(s.mux)
}

// Shutdown gracefully stops the Asynq server.
func (s *Server) Shutdown() {
	s.srv.Shutdown()
	s.client.Close()
}

// enqueueNext enqueues the next pipeline stage.
func (s *Server) enqueueNext(taskType string, payload pipeline.Payload) error {
	task, err := pipeline.NewTask(taskType, payload)
	if err != nil {
		return fmt.Errorf("enqueueNext: %w", err)
	}
	_, err = s.client.Enqueue(task)
	if err != nil {
		return fmt.Errorf("enqueueNext %s: %w", taskType, err)
	}
	return nil
}

// callLLM is a helper to call the LLM provider with retry.
func (s *Server) callLLM(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	messages := []adapter.LLMMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	opts := adapter.LLMOptions{
		MaxTokens:   maxTokens,
		Temperature: 0.7,
	}

	resp, err := s.llmProvider.Chat(ctx, messages, opts)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// ---------- Pipeline Stage Handlers ----------

func (s *Server) handleKeywordExpand(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: keyword expand", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageKeywordExpand, 10); err != nil {
		return err
	}

	// Get task details
	task, err := s.taskSvc.GetByID(ctx, p.TaskID)
	if err != nil || task == nil {
		return fmt.Errorf("handleKeywordExpand: task not found: %w", err)
	}

	// Call LLM to expand keyword into search queries
	content, err := s.callLLM(ctx,
		"你是一个内容研究助手。根据给定的关键词、受众和语气，生成5个搜索查询词，用于搜集相关素材。只返回JSON数组，不要其他内容。",
		fmt.Sprintf("关键词: %s\n受众: %s\n语气: %s\n\n请返回JSON数组格式的5个搜索查询词。", task.Keyword, task.Audience, task.Tone),
		500,
	)
	if err != nil {
		s.logger.Warn("LLM keyword expand failed, using original keyword", zap.Error(err))
		p.Queries = []string{task.Keyword}
	} else {
		// Parse JSON array
		var queries []string
		cleaned := extractJSON(content)
		if json.Unmarshal([]byte(cleaned), &queries) != nil || len(queries) == 0 {
			queries = []string{task.Keyword}
		}
		p.Queries = queries
	}

	s.logger.Info("keyword expand done", zap.Int("queries", len(p.Queries)))
	time.Sleep(500 * time.Millisecond) // Small delay for UI visibility
	return s.enqueueNext(pipeline.TypeSourceSearch, *p)
}

func (s *Server) handleSourceSearch(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: source search", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageSourceSearch, 20); err != nil {
		return err
	}

	// SearchProvider is still stub — simulate source results
	time.Sleep(1 * time.Second)
	return s.enqueueNext(pipeline.TypeContentCrawl, *p)
}

func (s *Server) handleContentCrawl(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: content crawl", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageContentCrawl, 35); err != nil {
		return err
	}

	// CrawlerProvider is still stub
	time.Sleep(1 * time.Second)
	return s.enqueueNext(pipeline.TypeHotScore, *p)
}

func (s *Server) handleHotScore(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: hot score", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageHotScore, 45); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)
	return s.enqueueNext(pipeline.TypeArticleWrite, *p)
}

func (s *Server) handleArticleWrite(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: article write", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageArticleWrite, 55); err != nil {
		return err
	}

	task, err := s.taskSvc.GetByID(ctx, p.TaskID)
	if err != nil || task == nil {
		return fmt.Errorf("handleArticleWrite: task not found")
	}

	// Call LLM to write the article
	content, err := s.callLLM(ctx,
		`你是一个专业的内容创作者。请根据要求撰写一篇高质量文章。
返回严格的JSON格式（不要markdown代码块），结构如下：
{
  "title": "文章标题",
  "digest": "100字以内的文章摘要",
  "blocks": [
    {"type": "lead", "content": "引言段落"},
    {"type": "section", "heading": "第一节标题", "content": "第一节内容..."},
    {"type": "section", "heading": "第二节标题", "content": "第二节内容..."},
    {"type": "section", "heading": "第三节标题", "content": "第三节内容..."},
    {"type": "cta", "content": "结尾行动号召"}
  ]
}`,
		fmt.Sprintf("关键词: %s\n目标受众: %s\n语气风格: %s\n目标字数: %d\n\n请撰写文章，确保内容原创、有深度、对读者有实际价值。", task.Keyword, task.Audience, task.Tone, task.TargetWords),
		4096,
	)
	if err != nil {
		s.taskSvc.MarkFailed(ctx, p.TaskID, fmt.Sprintf("AI 写作失败: %v", err))
		return err
	}

	// Parse the article JSON
	type articleBlock struct {
		Type    string `json:"type"`
		Heading string `json:"heading,omitempty"`
		Content string `json:"content"`
	}
	type articleOutput struct {
		Title  string         `json:"title"`
		Digest string         `json:"digest"`
		Blocks []articleBlock `json:"blocks"`
	}

	var article articleOutput
	cleaned := extractJSON(content)
	if err := json.Unmarshal([]byte(cleaned), &article); err != nil {
		// Fallback: create a simple draft with the raw content
		s.logger.Warn("failed to parse article JSON, using raw content", zap.Error(err))
		article = articleOutput{
			Title:  task.Keyword + " — AI 生成文章",
			Digest: "AI 自动生成的文章",
			Blocks: []articleBlock{
				{Type: "lead", Content: content},
			},
		}
	}

	if article.Title == "" {
		article.Title = task.Keyword
	}

	// Create draft in DB
	d := draft.ArticleDraft{
		TaskID:       task.ID,
		Title:        article.Title,
		Digest:       article.Digest,
		AuthorName:   "阅芽 AI",
		ReviewStatus: "pending",
		RiskLevel:    "low",
		Version:      1,
	}
	if err := s.draftRepo.Create(ctx, &d); err != nil {
		s.taskSvc.MarkFailed(ctx, p.TaskID, fmt.Sprintf("保存草稿失败: %v", err))
		return err
	}

	// Create blocks
	blocks := make([]draft.ArticleBlock, 0, len(article.Blocks))
	for i, ab := range article.Blocks {
		heading := ab.Heading
		textMD := ab.Content
		blocks = append(blocks, draft.ArticleBlock{
			DraftID:   d.ID,
			SortNo:    (i + 1) * 10,
			BlockType: ab.Type,
			Heading:   strPtr(heading),
			TextMD:    &textMD,
			Status:    "active",
		})
	}
	if len(blocks) > 0 {
		if err := s.blockRepo.CreateBatch(ctx, blocks); err != nil {
			s.logger.Error("failed to create blocks", zap.Error(err))
		}
	}

	// Link draft to task
	s.taskSvc.SetResultDraft(ctx, p.TaskID, d.ID)

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageArticleWrite, 70); err != nil {
		return err
	}

	return s.enqueueNext(pipeline.TypeImageMatch, *p)
}

func (s *Server) handleImageMatch(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: image match", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageImageMatch, 80); err != nil {
		return err
	}

	// Get the task to find the keyword for image search
	task, err := s.taskSvc.GetByID(ctx, p.TaskID)
	if err != nil || task == nil {
		s.logger.Warn("image match: task not found, skipping", zap.Int64("task_id", p.TaskID))
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	// Search for images using the keyword
	imgSvc := pipelinePkg.NewImageService(s.imageSearch, s.imageGen)
	results, err := imgSvc.SearchAndMatch(ctx, task.Keyword, 3)
	if err != nil {
		s.logger.Warn("image match: search failed, continuing without images", zap.Error(err))
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	if len(results) == 0 {
		s.logger.Info("image match: no images found")
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	// Get the draft and its blocks
	if task.ResultDraftID == nil {
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, *task.ResultDraftID)
	if err != nil || len(blocks) == 0 {
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	// Assign images to section blocks
	imgIdx := 0
	for i, block := range blocks {
		if block.BlockType == "section" && imgIdx < len(results) {
			imgTag := fmt.Sprintf(`<img src="%s" alt="%s" width="%d" style="width:100%%;border-radius:8px;margin:12px 0" />`,
				results[imgIdx].URL, task.Keyword, results[imgIdx].Width)

			existing := ""
			if block.HTMLFragment != nil {
				existing = *block.HTMLFragment
			}
			newHtml := imgTag + existing
			blocks[i].HTMLFragment = &newHtml
			s.blockRepo.Update(ctx, &blocks[i])

			imgIdx++
			if imgIdx >= 2 { // Max 2 images per article
				break
			}
		}
	}

	s.logger.Info("image match: assigned images", zap.Int("count", imgIdx))
	return s.enqueueNext(pipeline.TypeChartGen, *p)
}

func (s *Server) handleChartGen(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: chart gen", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageChartGen, 88); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)
	return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
}

func (s *Server) handleHTMLCompile(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: HTML compile", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageHTMLCompile, 95); err != nil {
		return err
	}

	// Load the draft and compile HTML
	task, _ := s.taskSvc.GetByID(ctx, p.TaskID)
	if task != nil && task.ResultDraftID != nil {
		d, _ := s.draftRepo.FindByID(ctx, *task.ResultDraftID)
		if d != nil {
			blocks, _ := s.blockRepo.FindByDraftID(ctx, d.ID)
			html := compileHTML(d.Title, blocks)
			d.CompiledHTML = html
			s.draftRepo.Update(ctx, d)
		}
	}

	// Mark task as done
	if err := s.taskSvc.MarkDone(ctx, p.TaskID); err != nil {
		return err
	}

	s.logger.Info("pipeline complete!", zap.Int64("task_id", p.TaskID))
	return nil
}

func (s *Server) handlePublish(ctx context.Context, t *asynq.Task) error {
	p, err := pipeline.ParsePayload(t)
	if err != nil {
		return err
	}

	s.logger.Info("pipeline: publish", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StagePublish, 98); err != nil {
		return err
	}

	// WeChat publisher is still stub
	return s.taskSvc.MarkDone(ctx, p.TaskID)
}

// ---------- Helpers ----------

func extractJSON(s string) string {
	// Strip markdown code fences if present
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
	}
	s = strings.TrimSpace(s)

	// Find first { or [
	start := -1
	for i, c := range s {
		if c == '{' || c == '[' {
			start = i
			break
		}
	}
	if start < 0 {
		return s
	}

	// Find matching closing bracket
	target := byte('}')
	if s[start] == '[' {
		target = ']'
	}
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == s[start] {
			depth++
		} else if s[i] == target {
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return s[start:]
}

func compileHTML(title string, blocks []draft.ArticleBlock) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<h1 style="font-size:24px;font-weight:bold;margin-bottom:16px;">%s</h1>`, title))

	for _, b := range blocks {
		switch b.BlockType {
		case draft.BlockTypeLead:
			text := derefStr(b.TextMD)
			sb.WriteString(fmt.Sprintf(`<p style="font-size:16px;color:#525252;margin-bottom:20px;line-height:1.8;">%s</p>`, text))
		case draft.BlockTypeSection:
			heading := derefStr(b.Heading)
			text := derefStr(b.TextMD)
			if heading != "" {
				sb.WriteString(fmt.Sprintf(`<h2 style="font-size:20px;font-weight:bold;margin:24px 0 12px;">%s</h2>`, heading))
			}
			sb.WriteString(fmt.Sprintf(`<p style="font-size:15px;line-height:1.8;margin-bottom:16px;">%s</p>`, text))
		case draft.BlockTypeCTA:
			text := derefStr(b.TextMD)
			sb.WriteString(fmt.Sprintf(`<p style="font-size:15px;color:#0a0a0a;font-weight:bold;margin-top:24px;padding:16px;background:#f5f5f5;border-radius:8px;">%s</p>`, text))
		default:
			text := derefStr(b.TextMD)
			if text != "" {
				sb.WriteString(fmt.Sprintf(`<p style="font-size:15px;line-height:1.8;margin-bottom:16px;">%s</p>`, text))
			}
		}
	}
	return sb.String()
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
