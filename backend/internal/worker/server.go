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
	"readbud/internal/domain/source"
	taskDomain "readbud/internal/domain/task"
	"readbud/internal/pipeline"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"
	pipelinePkg "readbud/internal/service/pipeline"
)

// Server wraps the Asynq server and registers pipeline handlers.
type Server struct {
	srv             *asynq.Server
	mux             *asynq.ServeMux
	client          *asynq.Client
	taskSvc         *service.TaskService
	draftRepo       postgres.ArticleDraftRepository
	blockRepo       postgres.ArticleBlockRepository
	sourceRepo      postgres.SourceDocumentRepository
	llmProvider     adapter.LLMProvider
	searchProvider  adapter.SearchProvider
	crawlerProvider adapter.CrawlerProvider
	imageSearch     adapter.ImageSearchProvider
	imageGen        adapter.ImageGenProvider
	logger          *zap.Logger
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
	searchProvider adapter.SearchProvider,
	crawlerProvider adapter.CrawlerProvider,
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
		srv:             srv,
		mux:             asynq.NewServeMux(),
		client:          client,
		taskSvc:         taskSvc,
		draftRepo:       draftRepo,
		blockRepo:       blockRepo,
		sourceRepo:      sourceRepo,
		llmProvider:     llmProvider,
		searchProvider:  searchProvider,
		crawlerProvider: crawlerProvider,
		imageSearch:     imageSearch,
		imageGen:        imageGen,
		logger:          logger,
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
	// When no article style is set, also ask LLM to recommend one
	if task.ArticleStyle == "" {
		type expandResult struct {
			Queries          []string `json:"queries"`
			RecommendedStyle string   `json:"recommended_style"`
		}

		systemPrompt := `你是一个内容研究助手。根据给定的关键词、受众和语气，完成两个任务：
1. 生成5个搜索查询词，用于搜集相关素材
2. 根据主题和受众特征，推荐最合适的文章风格

可选风格：
- minimal：极简专业型，适合知识、行业、B2B内容
- magazine：杂志编辑型，适合品牌故事、人物、深度观点
- listicle：清单干货型，适合教程、攻略、经验复盘
- narrative：叙事故事型，适合案例、转化、品牌温度
- faq：问答拆解型，适合教育用户、解释复杂概念
- casual：轻社交型，适合情绪、观点、生活方式

返回严格JSON（不要markdown代码块）：
{"queries": ["query1", "query2", "query3", "query4", "query5"], "recommended_style": "minimal"}`

		content, err := s.callLLM(ctx,
			systemPrompt,
			fmt.Sprintf("关键词: %s\n受众: %s\n语气: %s", task.Keyword, task.Audience, task.Tone),
			500,
		)
		if err != nil {
			s.logger.Warn("LLM keyword expand failed, using original keyword", zap.Error(err))
			p.Queries = []string{task.Keyword}
		} else {
			cleaned := extractJSON(content)
			var result expandResult
			if json.Unmarshal([]byte(cleaned), &result) != nil || len(result.Queries) == 0 {
				p.Queries = []string{task.Keyword}
			} else {
				p.Queries = result.Queries
				if taskDomain.ValidArticleStyles[result.RecommendedStyle] {
					if styleErr := s.taskSvc.UpdateArticleStyle(ctx, p.TaskID, result.RecommendedStyle); styleErr != nil {
						s.logger.Warn("failed to save recommended article style", zap.Error(styleErr))
					} else {
						s.logger.Info("recommended article style saved", zap.String("style", result.RecommendedStyle))
					}
				}
			}
		}
	} else {
		content, err := s.callLLM(ctx,
			"你是一个内容研究助手。根据给定的关键词、受众和语气，生成5个搜索查询词，用于搜集相关素材。只返回JSON数组，不要其他内容。",
			fmt.Sprintf("关键词: %s\n受众: %s\n语气: %s\n\n请返回JSON数组格式的5个搜索查询词。", task.Keyword, task.Audience, task.Tone),
			500,
		)
		if err != nil {
			s.logger.Warn("LLM keyword expand failed, using original keyword", zap.Error(err))
			p.Queries = []string{task.Keyword}
		} else {
			var queries []string
			cleaned := extractJSON(content)
			if json.Unmarshal([]byte(cleaned), &queries) != nil || len(queries) == 0 {
				queries = []string{task.Keyword}
			}
			p.Queries = queries
		}
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

	// Use expanded queries from previous stage (or original keyword)
	queries := p.Queries
	if len(queries) == 0 {
		task, _ := s.taskSvc.GetByID(ctx, p.TaskID)
		if task != nil {
			queries = []string{task.Keyword}
		}
	}

	// Search using all queries, deduplicate by URL
	seen := make(map[string]bool)
	var allResults []adapter.SearchResult
	for _, q := range queries {
		results, searchErr := s.searchProvider.Search(ctx, q, adapter.SearchOptions{MaxResults: 3})
		if searchErr != nil {
			s.logger.Warn("source search: query failed", zap.String("query", q), zap.Error(searchErr))
			continue
		}
		for _, r := range results {
			if !seen[r.URL] {
				seen[r.URL] = true
				allResults = append(allResults, r)
			}
		}
	}

	s.logger.Info("source search: found sources", zap.Int("count", len(allResults)))

	// Save sources to DB
	task, _ := s.taskSvc.GetByID(ctx, p.TaskID)
	if task != nil {
		now := time.Now()
		for i, r := range allResults {
			if i >= 10 {
				break
			}
			src := source.SourceDocument{
				TaskID:     task.ID,
				SourceType: source.SourceTypeWeb,
				SourceURL:  r.URL,
				Title:      r.Title,
				CrawledAt:  now,
			}
			if createErr := s.sourceRepo.Create(ctx, &src); createErr != nil {
				s.logger.Warn("source search: failed to save source", zap.String("url", r.URL), zap.Error(createErr))
			}
		}
	}

	// Store URLs in payload for crawl stage
	var urls []string
	for _, r := range allResults {
		urls = append(urls, r.URL)
		if len(urls) >= 5 {
			break
		}
	}
	p.SourceURLs = urls

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

	// Crawl each source URL
	var crawledContents []string
	for _, u := range p.SourceURLs {
		page, crawlErr := s.crawlerProvider.Crawl(ctx, u)
		if crawlErr != nil {
			s.logger.Warn("content crawl: failed", zap.String("url", u), zap.Error(crawlErr))
			continue
		}
		crawledContents = append(crawledContents, fmt.Sprintf("--- Source: %s ---\n%s", page.Title, page.Content))
	}

	s.logger.Info("content crawl: crawled", zap.Int("count", len(crawledContents)))

	// Store crawled content in payload for later stages
	p.CrawledContent = strings.Join(crawledContents, "\n\n")

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

	if p.CrawledContent == "" {
		s.logger.Info("hot score: no crawled content, skipping")
		return s.enqueueNext(pipeline.TypeArticleWrite, *p)
	}

	// Use LLM to analyze and rank the crawled content
	task, _ := s.taskSvc.GetByID(ctx, p.TaskID)
	keyword := ""
	if task != nil {
		keyword = task.Keyword
	}

	scored, llmErr := s.callLLM(ctx,
		`你是一个内容分析师。分析以下搜集到的素材，评估每段内容与主题的相关性、信息质量和独特性。
返回一段精炼的、高质量的素材摘要（不超过2000字），只保留最有价值的信息、数据、观点和案例。
去除重复内容、广告、无关信息。直接返回整理后的文本，不需要JSON格式。`,
		fmt.Sprintf("主题关键词: %s\n\n搜集到的素材:\n%s", keyword, p.CrawledContent),
		3000,
	)
	if llmErr != nil {
		s.logger.Warn("hot score: LLM analysis failed, using raw content", zap.Error(llmErr))
		// Keep raw content as-is
	} else {
		p.CrawledContent = scored
	}

	s.logger.Info("hot score: analysis done", zap.Int("content_len", len(p.CrawledContent)))
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
		`你是一位拥有十年以上写作经验的资深内容专家，擅长将复杂话题转化为引人入胜、富有洞见的深度文章。你的文字风格兼具专业深度与阅读愉悦感——像一位博学的朋友在跟读者促膝长谈。

写作原则：
- 观点鲜明：每个段落都有明确的核心论点，不说正确的废话
- 有血有肉：用真实的案例、数据、场景描写来支撑观点，避免空洞说教
- 节奏感强：长短句交替，适当使用反问、设问、类比来制造阅读节奏
- 反常识切入：开头尝试用一个颠覆认知的事实、一个反直觉的观点、或一个具体的场景切入，而非老套的"随着...的发展"
- 自然表达：像真人在写文章，允许偶尔的口语化表达、个人判断和态度，而非面面俱到的中立客观
- 结尾有力：结尾给读者一个可执行的行动建议或一个值得深思的问题，而非空洞总结

绝对禁止的 AI 味道：
- 不要使用"让我们""在当今社会""随着...的发展""总而言之""综上所述"等套话
- 不要每段都以"首先""其次""最后"来结构化
- 不要使用排比句超过2组
- 不要在结尾做大而空的升华

同时，请在文章开头的 lead block 中写一段能立即抓住注意力的开场白（可以是一个震撼数据、一个反常识断言、或一个读者一定有共鸣的场景描写），让人忍不住继续读下去。

在结尾的 cta block 中，设计一段有创意的结尾：不要简单的"关注我们"，而是给出一个具体的、读者今天就能做的小行动，或抛出一个引发思考的好问题。用精美的HTML来呈现这段结尾卡片，使其具有视觉吸引力。

返回严格的JSON格式（不要markdown代码块），结构如下：
{
  "title": "一个让人想点开的标题（可以用冒号分隔主副标题）",
  "digest": "100字以内的文章摘要，要有信息密度，不要空洞形容词",
  "cover_prompt": "一段英文提示词，用于AI生成与文章主题匹配的封面图，描述清晰、色彩明亮、构图优美",
  "blocks": [
    {"type": "lead", "content": "引言段落（抓眼球的开头）"},
    {"type": "section", "heading": "第一节标题", "content": "第一节正文内容，支持HTML标签如<strong>加粗</strong>、<em>斜体</em>等"},
    {"type": "section", "heading": "第二节标题", "content": "第二节内容..."},
    {"type": "section", "heading": "第三节标题", "content": "第三节内容..."},
    {"type": "cta", "content": "<div style='background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);border-radius:16px;padding:32px;color:#fff;text-align:center'><h3 style='margin:0 0 12px;font-size:20px'>结尾卡片标题</h3><p style='margin:0;opacity:0.9;font-size:15px;line-height:1.6'>具体的行动号召或思考问题</p></div>"}
  ]
}`,
		func() string {
			userPrompt := fmt.Sprintf("关键词: %s\n目标受众: %s\n语气风格: %s\n目标字数: %d", task.Keyword, task.Audience, task.Tone, task.TargetWords)
			if p.CrawledContent != "" {
				userPrompt += fmt.Sprintf("\n\n以下是搜集到的高质量参考素材，请基于这些素材撰写文章（但不要照搬，要用自己的语言重新组织）：\n\n%s", p.CrawledContent)
			}
			userPrompt += "\n\n请开始撰写。记住：写出真正有价值、有洞见、让人读完有收获的好文章，像一位行业专家在分享真知灼见，而不是AI在堆砌信息。"
			return userPrompt
		}(),
		8192,
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
		Title       string         `json:"title"`
		Digest      string         `json:"digest"`
		CoverPrompt string         `json:"cover_prompt,omitempty"`
		Blocks      []articleBlock `json:"blocks"`
	}

	var article articleOutput
	cleaned := extractJSON(content)
	if err := json.Unmarshal([]byte(cleaned), &article); err != nil {
		// Fallback: create a simple draft with the raw content
		s.logger.Warn("failed to parse article JSON, using raw content", zap.Error(err))
		article = articleOutput{
			Title:  task.Keyword,
			Digest: "由阅芽内容引擎创作",
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

	// Translate keyword to English for better Pexels results
	searchQuery := task.Keyword
	engQuery, engErr := s.callLLM(ctx,
		"Translate the following keyword to a short English search phrase suitable for stock photo search. Return ONLY the English phrase, nothing else.",
		task.Keyword,
		50,
	)
	if engErr == nil && len(engQuery) > 0 && len(engQuery) < 100 {
		searchQuery = engQuery
		s.logger.Info("image match: translated query", zap.String("original", task.Keyword), zap.String("english", searchQuery))
	}

	// Search for images
	imgSvc := pipelinePkg.NewImageService(s.imageSearch, s.imageGen)
	results, err := imgSvc.SearchAndMatch(ctx, searchQuery, 3)
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
			imgTag := fmt.Sprintf(`<figure style="margin:20px 0;text-align:center"><img src="%s" alt="%s" style="width:100%%;max-width:100%%;border-radius:12px;display:block" /><figcaption style="font-size:12px;color:#999;margin-top:8px">图片来源: Pexels</figcaption></figure>`,
				results[imgIdx].URL, task.Keyword)

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

	s.logger.Info("pipeline: visual enhance (chart gen stage)", zap.Int64("task_id", p.TaskID))
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageChartGen, 88); err != nil {
		return err
	}

	task, err := s.taskSvc.GetByID(ctx, p.TaskID)
	if err != nil || task == nil || task.ResultDraftID == nil {
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	blocks, err := s.blockRepo.FindByDraftID(ctx, *task.ResultDraftID)
	if err != nil || len(blocks) == 0 {
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	// Build current article summary for LLM context
	var articleSummary strings.Builder
	for i, b := range blocks {
		bType := b.BlockType
		heading := derefStr(b.Heading)
		text := derefStr(b.TextMD)
		hasImg := b.HTMLFragment != nil && strings.Contains(*b.HTMLFragment, "<img")
		articleSummary.WriteString(fmt.Sprintf("Block %d [%s]%s: %s\n", i, bType, func() string {
			if heading != "" { return " heading=\"" + heading + "\"" }
			return ""
		}(), func() string {
			if len(text) > 150 { return text[:150] + "..." }
			return text
		}()))
		if hasImg {
			articleSummary.WriteString("  (已有配图)\n")
		}
	}

	// LLM visual enhancement round
	// Randomly select a visual style for variety
	visualStyles := []string{
		"有机呼吸风：柔和的blob形状、手绘感曲线、粉紫青渐变、轻雾面质感。形状如有机物在缓慢呼吸，边缘柔软模糊。",
		"极简科技风：深色卡片+霓虹描边、精密几何SVG图标、蓝紫渐变辉光、数据面板质感。冷峻但不冰冷。",
		"自然手账风：米色/暖白底色、水彩渐变色块、圆润胶囊标签、手写体标题装饰线。温暖亲切如纸质手账。",
		"品牌杂志风：大胆撞色、粗体序号、几何色块分区、全宽色带标题栏。像一本精心排版的生活方式杂志。",
		"学术优雅风：深蓝/墨绿主色、精致细线框、serif风格序号、脚注式引用卡片。严谨又不失美感。",
		"创意工作室风：不规则圆角卡片、渐变色投影、SVG波浪分隔线、略带倾斜的标签元素。活泼但不幼稚。",
		"赛博朋克风：暗底+荧光绿/粉/蓝高亮、终端代码框样式、扫描线纹理叠层、数据流视觉隐喻。未来感十足。",
	}
	// Use task ID for deterministic but varied selection
	styleIdx := int(p.TaskID) % len(visualStyles)
	selectedStyle := visualStyles[styleIdx]

	enhancePrompt := fmt.Sprintf(`你是一位顶级的微信公众号视觉排版设计师，拥有极高的审美品味。你的使命是将普通文章变成让人忍不住截图分享的精美作品。

本次设计风格指令：%s

## 核心设计原则

### 图标系统（严禁使用emoji）
所有图标必须使用内联SVG。为每个场景设计恰当的SVG图标：
- 序号标识：用主题色圆形/圆角矩形包裹数字，或用svg绘制创意序号
- 段落图标：根据内容含义用svg绘制简洁的示意图标（如灯泡表示观点、图表表示数据、箭头表示趋势等）
- 装饰图标：svg绘制的小装饰元素（星星、圆点、曲线等）
- SVG保持简洁：每个图标不超过3-4个path，viewBox统一用"0 0 24 24"，stroke-width用"1.5"或"2"

### 视觉材质
- 多值圆角（border-radius: 20px 8px 20px 8px 这样的不规则圆角）
- 柔和渐变（linear-gradient 使用2-3个相近色，角度135deg或180deg）
- 轻透明叠层（rgba背景 + 细描边）
- 阴影用多层柔和投影（0 2px 12px rgba(...)）而非单层硬阴影
- 绝对不要用1px solid #000这种生硬线条

### 动画（微信公众号支持CSS animation）
在关键元素上添加微动画：
- 卡片：悬浮感（@keyframes float { 0%%,100%% { transform:translateY(0) } 50%% { transform:translateY(-4px) } }），6-10秒周期
- 渐变色块：色彩缓慢流动（background-size:200%% + @keyframes gradient-shift）
- 装饰元素：缓慢旋转或脉冲（@keyframes pulse-soft）
- 注意：animation写在inline style里，@keyframes用<style>标签包裹放在HTML开头

### 版式节奏
- lead开头：品牌装饰条（渐变色svg波浪线或自定义形状）+ 精排首段
- 每个section：创意序号 + 标题装饰（如背景色块、下划线动画、侧边装饰条）
- 正文中：选择性高亮（渐变背景的inline highlight，不要全部加粗）
- 重要观点：独立卡片展示（圆角+渐变背景+svg图标+投影）
- 数据/对比：设计信息卡片或小型数据面板
- 每2-3段之间：插入一个视觉呼吸元素（svg波浪分隔线、创意引用框、tips卡片等）
- CTA结尾：设计一个精美的行动号召卡片——使用曲线clip-path裁剪的形状、blob渐变背景、柔和投影

### 配色
根据文章主题和选定风格，自创一组和谐配色（主色+辅色+点缀色+背景色）。不要使用固定的蓝紫粉——每篇文章的配色都应该不同且与内容呼应。

## 技术约束
- 所有样式使用inline style（微信不支持class）
- @keyframes动画放在开头的<style>标签内（这是唯一允许的非inline样式）
- 不要使用JavaScript
- 不要改变文字内容和含义
- 已有的<img>和<figure>标签保留原样
- SVG直接内联在HTML中，不要用外部引用

## 输出格式
返回严格JSON（不要markdown代码块）：
{
  "style_name": "风格名称",
  "color_scheme": {"primary": "#xxx", "secondary": "#xxx", "accent": "#xxx", "bg": "#xxx"},
  "animation_css": "<style>这里放所有@keyframes定义</style>",
  "blocks": [
    {"index": 0, "html": "该block的完整HTML（包含原始文字+所有装饰元素+SVG图标+动画）"},
    {"index": 1, "html": "..."}
  ]
}`, selectedStyle)

	enhanceContent, err := s.callLLM(ctx, enhancePrompt, articleSummary.String(), 8192)
	if err != nil {
		s.logger.Warn("visual enhance failed, continuing with original content", zap.Error(err))
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	// Parse enhancement result
	type enhanceBlock struct {
		Index int    `json:"index"`
		HTML  string `json:"html"`
	}
	type enhanceColorScheme struct {
		Primary   string `json:"primary"`
		Secondary string `json:"secondary"`
		Accent    string `json:"accent"`
		Bg        string `json:"bg"`
	}
	type enhanceOutput struct {
		StyleName    string              `json:"style_name"`
		ColorScheme  json.RawMessage     `json:"color_scheme"`
		AnimationCSS string              `json:"animation_css"`
		Blocks       []enhanceBlock      `json:"blocks"`
	}

	var enhanced enhanceOutput
	cleaned := extractJSON(enhanceContent)
	if err := json.Unmarshal([]byte(cleaned), &enhanced); err != nil {
		s.logger.Warn("visual enhance: failed to parse LLM output", zap.Error(err))
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	// Apply enhanced HTML to blocks
	applied := 0

	// Prepend animation CSS to the first block if provided
	animCSS := enhanced.AnimationCSS
	firstBlockDone := false

	for _, eb := range enhanced.Blocks {
		if eb.Index >= 0 && eb.Index < len(blocks) && eb.HTML != "" {
			// Preserve existing images
			existingHTML := ""
			if blocks[eb.Index].HTMLFragment != nil {
				existingHTML = *blocks[eb.Index].HTMLFragment
			}

			// If existing has images, merge them into the new HTML
			newHTML := eb.HTML
			if strings.Contains(existingHTML, "<figure") || strings.Contains(existingHTML, "<img") {
				imgStart := strings.Index(existingHTML, "<figure")
				if imgStart == -1 {
					imgStart = strings.Index(existingHTML, "<img")
				}
				if imgStart >= 0 {
					imgEnd := strings.Index(existingHTML[imgStart:], "</figure>")
					if imgEnd >= 0 {
						imgTag := existingHTML[imgStart : imgStart+imgEnd+len("</figure>")]
						newHTML = imgTag + newHTML
					}
				}
			}

			// Prepend animation CSS to the very first enhanced block
			if !firstBlockDone && animCSS != "" {
				newHTML = animCSS + newHTML
				firstBlockDone = true
			}

			blocks[eb.Index].HTMLFragment = &newHTML
			s.blockRepo.Update(ctx, &blocks[eb.Index])
			applied++
		}
	}

	s.logger.Info("visual enhance: applied",
		zap.Int("blocks_enhanced", applied),
		zap.String("style", enhanced.StyleName))
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
		// If html_fragment exists (from visual enhance), use it directly
		if b.HTMLFragment != nil && *b.HTMLFragment != "" {
			sb.WriteString(*b.HTMLFragment)
			continue
		}

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
