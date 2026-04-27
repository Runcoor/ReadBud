// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package worker

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain"
	"readbud/internal/domain/asset"
	"readbud/internal/domain/draft"
	"readbud/internal/domain/source"
	taskDomain "readbud/internal/domain/task"
	"readbud/internal/pipeline"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"
	"readbud/internal/service/imageresize"
	pipelinePkg "readbud/internal/service/pipeline"

	"gorm.io/datatypes"
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
	brandRepo       postgres.BrandProfileRepository
	llmProvider     adapter.LLMProvider
	searchProvider  adapter.SearchProvider
	crawlerProvider adapter.CrawlerProvider
	imageSearch     adapter.ImageSearchProvider
	imageGen        adapter.ImageGenProvider
	assetRepo       postgres.AssetRepository
	storage         adapter.StorageProvider
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
	brandRepo postgres.BrandProfileRepository,
	llmProvider adapter.LLMProvider,
	searchProvider adapter.SearchProvider,
	crawlerProvider adapter.CrawlerProvider,
	imageSearch adapter.ImageSearchProvider,
	imageGen adapter.ImageGenProvider,
	assetRepo postgres.AssetRepository,
	storageProvider adapter.StorageProvider,
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
		brandRepo:       brandRepo,
		llmProvider:     llmProvider,
		searchProvider:  searchProvider,
		crawlerProvider: crawlerProvider,
		imageSearch:     imageSearch,
		imageGen:        imageGen,
		assetRepo:       assetRepo,
		storage:         storageProvider,
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

// ---------- Prompt Builder Helpers ----------

var styleSkeletonMap = map[string]string{
	"minimal": `【本次预设：极简专业型】
- 设计基因：黑白底色 + 荧光黄高亮，衬线标题 + 等宽编号
- 开头方式：用一个震撼数据或反常识断言切入；首句尽量保留可被高亮的短语（用 ==…== 标注）
- 结构：lead → 3-5 个 section（每节 H2 会自动加 mono 编号 01/02/03）→ checklist 或 summary → cta
- 段落：短段落，高信息密度，每段 1-3 行
- 小标题：简短有力，4-12 个字
- 可用 block：lead, section, quote, checklist, summary, cta`,

	"magazine": `【本次预设：杂志编辑型】
- 设计基因：米色纸张底 + 报刊红强调，Bodoni 大字 + 报头报尾
- 开头方式：用一个具体的场景描写或人物故事切入；首段会渲染为红色首字下沉，请确保首字有意义
- 结构：lead（首段长一点，至少 3-4 句）→ 4-5 个 section（穿插 quote 金句、可选 summary 作为 FINAL TAKE）→ cta
- 段落：中长段落，叙事感强，画面感强
- 小标题：有编辑品味，配合"红色横条 + 大号衬线"，可双行
- 可用 block：lead, section, quote, summary, cta`,

	"stitch": `【本次预设：暖橙手账型】
- 设计基因：米色底 + 暖橙强调，居中标题 + 装饰短横
- 开头方式：用一个轻松的场景或自问切入；lead 段落会被渲染为橙色软底卡片，文字温度要够
- 结构：lead → 3-5 个 section（H2 居中、配橙色短横）→ checklist 或 quote → cta
- 段落：节奏舒缓，长短交替；金句单独成段（用 ==…== 标注让其自动渲染为橙色加粗）
- 小标题：可以是观点句，5-14 个字，避免硬冷术语
- 可用 block：lead, section, quote, checklist, cta`,
}

func buildStyleSkeleton(style string) string {
	if s, ok := styleSkeletonMap[style]; ok {
		return s
	}
	return styleSkeletonMap["minimal"]
}

func (s *Server) buildBrandConstraint(ctx context.Context, brandProfileID *int64) string {
	var bp *domain.BrandProfile

	if brandProfileID != nil {
		bp, _ = s.brandRepo.FindByID(ctx, *brandProfileID)
	}
	if bp == nil {
		bp, _ = s.brandRepo.FindDefault(ctx)
	}
	if bp == nil {
		return "【品牌调性】专业、有温度、简洁\n【禁用词】让我们、随着...的发展、赋能、闭环、打造、值得一提的是、综上所述、在当今社会、毋庸置疑\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("【品牌调性】%s\n", bp.BrandTone))

	var forbidden []string
	if len(bp.ForbiddenWords) > 0 {
		_ = json.Unmarshal(bp.ForbiddenWords, &forbidden)
	}
	if len(forbidden) == 0 {
		forbidden = []string{"让我们", "随着...的发展", "赋能", "闭环", "打造", "值得一提的是", "综上所述", "在当今社会", "毋庸置疑"}
	}
	sb.WriteString(fmt.Sprintf("【禁用词】%s\n", strings.Join(forbidden, "、")))

	var preferred []string
	if len(bp.PreferredWords) > 0 {
		_ = json.Unmarshal(bp.PreferredWords, &preferred)
	}
	if len(preferred) > 0 {
		sb.WriteString(fmt.Sprintf("【偏好用词】%s\n", strings.Join(preferred, "、")))
	}

	return sb.String()
}

func (s *Server) buildDedupConstraint(ctx context.Context, brandProfileID *int64) string {
	drafts, err := s.draftRepo.FindRecentFingerprints(ctx, 10, brandProfileID)
	if err != nil || len(drafts) == 0 {
		return ""
	}

	openingCounts := map[string]int{}
	titleCounts := map[string]int{}
	ctaCounts := map[string]int{}

	for _, d := range drafts {
		if d.OpeningType != "" {
			openingCounts[d.OpeningType]++
		}
		if d.TitlePattern != "" {
			titleCounts[d.TitlePattern]++
		}
		if d.CTAType != "" {
			ctaCounts[d.CTAType]++
		}
	}

	var sb strings.Builder
	sb.WriteString("【变化要求——避免与近期文章重复】\n")

	if len(openingCounts) > 0 {
		sb.WriteString("- 近期开头方式已用过：")
		parts := []string{}
		for k, v := range openingCounts {
			parts = append(parts, fmt.Sprintf("%s(%d次)", k, v))
		}
		sb.WriteString(strings.Join(parts, "、"))
		allOpenings := []string{"data", "scene", "contrarian", "question", "story"}
		unused := []string{}
		for _, o := range allOpenings {
			if openingCounts[o] == 0 {
				unused = append(unused, o)
			}
		}
		if len(unused) > 0 {
			sb.WriteString("，请优先使用：" + strings.Join(unused, "、"))
		}
		sb.WriteString("\n")
	}

	if len(titleCounts) > 0 {
		sb.WriteString("- 近期标题句式已用过：")
		parts := []string{}
		for k, v := range titleCounts {
			parts = append(parts, fmt.Sprintf("%s(%d次)", k, v))
		}
		sb.WriteString(strings.Join(parts, "、"))
		allTitles := []string{"question", "colon", "number", "assertion", "howto"}
		unused := []string{}
		for _, t := range allTitles {
			if titleCounts[t] == 0 {
				unused = append(unused, t)
			}
		}
		if len(unused) > 0 {
			sb.WriteString("，请优先使用：" + strings.Join(unused, "、"))
		}
		sb.WriteString("\n")
	}

	if len(ctaCounts) > 0 {
		sb.WriteString("- 近期CTA方式已用过：")
		parts := []string{}
		for k, v := range ctaCounts {
			parts = append(parts, fmt.Sprintf("%s(%d次)", k, v))
		}
		sb.WriteString(strings.Join(parts, "、"))
		allCTAs := []string{"action", "question", "challenge", "reflection"}
		unused := []string{}
		for _, c := range allCTAs {
			if ctaCounts[c] == 0 {
				unused = append(unused, c)
			}
		}
		if len(unused) > 0 {
			sb.WriteString("，请优先使用：" + strings.Join(unused, "、"))
		}
		sb.WriteString("\n")
	}

	return sb.String()
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
2. 根据主题和受众特征，从下面三种排版预设里挑一个最合适的

可选预设（仅这三种）：
- minimal：极简专业型。黑白底色 + 荧光黄高亮，衬线标题 + 等宽编号，适合技术、AI、产品、知识类深度内容。
- magazine：杂志编辑型。米色纸张底 + 报刊红强调，Bodoni 大字 + 报头报尾，适合品牌故事、人物专访、文化、深度观点。
- stitch：暖橙手账型。米色底 + 暖橙强调，居中标题 + 短装饰横线，适合教程、生活方式、观点分享、轻量科普。

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

	// Build 5-module system prompt
	var sysPrompt strings.Builder

	// Module 1: Role
	sysPrompt.WriteString("你是一位资深微信公众号主编，擅长把复杂话题写成\"简洁、有质感、有传播力\"的图文文章。你的文字像一位博学的朋友在促膝长谈——专业但不居高临下，有态度但不偏激。\n\n")

	// Module 2: Brand constraint
	sysPrompt.WriteString(s.buildBrandConstraint(ctx, task.BrandProfileID))
	sysPrompt.WriteString("\n")

	// Module 3: Style skeleton
	style := task.ArticleStyle
	if style == "" {
		style = "minimal"
	}
	sysPrompt.WriteString(buildStyleSkeleton(style))
	sysPrompt.WriteString("\n\n")

	// Module 4: Dedup constraint
	dedup := s.buildDedupConstraint(ctx, task.BrandProfileID)
	if dedup != "" {
		sysPrompt.WriteString(dedup)
		sysPrompt.WriteString("\n")
	}

	// Module 5: Writing rules + output format
	sysPrompt.WriteString(`【写作规范】
- 观点鲜明：每段都有核心论点，不说正确的废话
- 有血有肉：用真实案例、数据、场景描写支撑观点，不空洞说教
- 节奏感：长短句交替，适当用反问、类比制造阅读节奏
- 自然表达：像真人写文章，允许口语化和个人判断
- 手机阅读：段落短，每段1-3行，留白足，小标题清晰
- 排比句不超过2组
- 不要每段都用"首先""其次""最后"结构化
- 结尾给读者一个可执行的行动建议或值得深思的问题

【输出格式】返回严格JSON（不要markdown代码块）：
{
  "titles": ["标题备选1", "标题备选2", "标题备选3"],
  "digest": "100字以内摘要，有信息密度，不要空洞形容词",
  "cover_prompt": "英文封面图提示词，描述清晰、色彩明亮、构图优美",
  "opening_type": "本文使用的开头方式：data|scene|contrarian|question|story",
  "title_pattern": "推荐的标题句式：question|colon|number|assertion|howto",
  "cta_type": "结尾方式：action|question|challenge|reflection",
  "blocks": [
    {"type": "lead", "content": "引言段落（抓眼球的开头）"},
    {"type": "section", "heading": "小标题", "content": "正文内容"},
    {"type": "quote", "content": "金句或重要引用"},
    {"type": "checklist", "content": "- 要点1\n- 要点2\n- 要点3"},
    {"type": "summary", "content": "核心结论浓缩"},
    {"type": "cta", "content": "行动号召或互动问题"}
  ]
}

注意：blocks中请根据风格要求灵活选用block类型，不需要全部使用。titles必须提供3个不同句式的备选标题。`)

	// Call LLM to write the article
	content, err := s.callLLM(ctx,
		sysPrompt.String(),
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
		Titles       []string       `json:"titles"`
		Title        string         `json:"title"`
		Digest       string         `json:"digest"`
		CoverPrompt  string         `json:"cover_prompt,omitempty"`
		OpeningType  string         `json:"opening_type,omitempty"`
		TitlePattern string         `json:"title_pattern,omitempty"`
		CTAType      string         `json:"cta_type,omitempty"`
		Blocks       []articleBlock `json:"blocks"`
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

	title := article.Title
	if len(article.Titles) > 0 {
		title = article.Titles[0]
	}
	if title == "" {
		title = task.Keyword
	}

	// Create draft in DB
	d := draft.ArticleDraft{
		TaskID:       task.ID,
		Title:        title,
		Digest:       article.Digest,
		AuthorName:   "阅芽 AI",
		ReviewStatus: "pending",
		RiskLevel:    "low",
		Version:      1,
		StyleUsed:    style,
		OpeningType:  article.OpeningType,
		TitlePattern: article.TitlePattern,
		CTAType:      article.CTAType,
	}
	if len(article.Titles) > 1 {
		titlesJSON, _ := json.Marshal(map[string]interface{}{
			"title_options": article.Titles,
		})
		d.OutlineJSON = datatypes.JSON(titlesJSON)
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

// matchedImage carries both the embed-ready URL and the persisted asset ID
// so callers can both render the article and link the asset to a block.
type matchedImage struct {
	URL     string
	AssetID int64
}

// persistImageAsAsset writes raw image bytes to storage AND creates an asset
// row, so the WeChat publish flow can later find and re-upload the image.
//
//	bucket        — storage bucket ("generated" or "search")
//	data          — raw image bytes
//	sourceKind    — asset.SourceKindGenerated or asset.SourceKindSearch
//	isAIGenerated — 1 for image-gen output, 0 for stock search
//	sourceURL     — original URL (Pexels page) for search; nil for generated
//	promptText    — the LLM prompt used to generate; nil for search
func (s *Server) persistImageAsAsset(
	ctx context.Context,
	bucket string,
	data []byte,
	sourceKind string,
	isAIGenerated int16,
	sourceURL *string,
	promptText *string,
) (string, int64, error) {
	// WeChat content image API caps each image at 1 MB. Fit the bytes
	// before they touch storage so the asset row's size matches what we
	// will actually upload later. Fitting may transcode PNG → JPEG.
	origSize := len(data)
	fitted, fittedMime, fitErr := imageresize.FitWeChat(data, adapter.ContentImageMaxBytes)
	if fitErr != nil {
		s.logger.Warn("image resize failed; persisting original",
			zap.String("bucket", bucket),
			zap.Int("orig_size", origSize),
			zap.Error(fitErr),
		)
	} else {
		if len(fitted) != origSize {
			s.logger.Info("image fitted to wechat content image limit",
				zap.String("bucket", bucket),
				zap.Int("orig_size", origSize),
				zap.Int("new_size", len(fitted)),
				zap.String("new_mime", fittedMime),
			)
		}
		data = fitted
	}

	mime := http.DetectContentType(data)
	ext := "png"
	switch mime {
	case "image/png":
		ext = "png"
	case "image/jpeg":
		ext = "jpg"
	default:
		return "", 0, fmt.Errorf("persistImageAsAsset: unsupported mime %q", mime)
	}

	now := time.Now().UTC()
	id := ulid.MustNew(ulid.Timestamp(now), crand.Reader).String()
	key := fmt.Sprintf("%04d%02d/%s.%s", now.Year(), now.Month(), id, ext)

	publicURL, err := s.storage.Upload(ctx, bucket, key, data, mime)
	if err != nil {
		return "", 0, fmt.Errorf("persistImageAsAsset: upload: %w", err)
	}

	sum := sha256.Sum256(data)
	sha := hex.EncodeToString(sum[:])
	sz := int64(len(data))

	assetType := asset.AssetTypeContentImage
	if isAIGenerated == 1 {
		assetType = asset.AssetTypeGeneratedImage
	}

	a := &asset.Asset{
		AssetType:          assetType,
		SourceKind:         sourceKind,
		MimeType:           mime,
		StorageProvider:    "local",
		Bucket:             bucket,
		ObjectKey:          key,
		SizeBytes:          &sz,
		SHA256:             sha,
		SourceURL:          sourceURL,
		PromptText:         promptText,
		IsAIGenerated:      isAIGenerated,
		WechatUploadStatus: asset.WechatUploadPending,
	}
	if err := s.assetRepo.Create(ctx, a); err != nil {
		// Best-effort rollback: remove the orphan file. Use Background context
		// because the original ctx may already be cancelled.
		_ = s.storage.Delete(context.Background(), bucket, key)
		return "", 0, fmt.Errorf("persistImageAsAsset: create asset: %w", err)
	}
	return publicURL, a.ID, nil
}

// generateImageWithAsset runs imageGen, downloads/decodes bytes, persists them
// via persistImageAsAsset, and returns a matchedImage ready for embedding.
func (s *Server) generateImageWithAsset(ctx context.Context, prompt string) (*matchedImage, error) {
	gen, err := s.imageGen.Generate(ctx, prompt, adapter.ImageGenOptions{
		Width:  1024,
		Height: 768,
		Style:  "professional",
	})
	if err != nil {
		return nil, fmt.Errorf("generateImageWithAsset: gen: %w", err)
	}

	var data []byte
	switch {
	case gen.Base64 != "":
		raw := gen.Base64
		if i := strings.Index(raw, ","); i >= 0 && strings.Contains(raw[:i], "base64") {
			raw = raw[i+1:]
		}
		decoded, decErr := base64.StdEncoding.DecodeString(raw)
		if decErr != nil {
			return nil, fmt.Errorf("generateImageWithAsset: decode base64: %w", decErr)
		}
		data = decoded
	case gen.URL != "":
		// Some OpenAI-compatible relays return the image inline as a data URL
		// instead of a public https:// URL. Detect and decode that case.
		if strings.HasPrefix(gen.URL, "data:") {
			raw := gen.URL
			if i := strings.Index(raw, ","); i >= 0 && strings.Contains(raw[:i], "base64") {
				raw = raw[i+1:]
			} else {
				return nil, fmt.Errorf("generateImageWithAsset: data URL not base64-encoded")
			}
			decoded, decErr := base64.StdEncoding.DecodeString(raw)
			if decErr != nil {
				return nil, fmt.Errorf("generateImageWithAsset: decode data URL: %w", decErr)
			}
			data = decoded
			break
		}
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, gen.URL, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("generateImageWithAsset: build request: %w", reqErr)
		}
		client := &http.Client{Timeout: 30 * time.Second}
		resp, getErr := client.Do(req)
		if getErr != nil {
			return nil, fmt.Errorf("generateImageWithAsset: download: %w", getErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("generateImageWithAsset: download status %d", resp.StatusCode)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
		if readErr != nil {
			return nil, fmt.Errorf("generateImageWithAsset: read body: %w", readErr)
		}
		data = body
	default:
		return nil, fmt.Errorf("generateImageWithAsset: empty result (no URL, no base64)")
	}

	promptCopy := prompt
	publicURL, assetID, err := s.persistImageAsAsset(
		ctx,
		"generated",
		data,
		asset.SourceKindGenerated,
		1,
		nil,
		&promptCopy,
	)
	if err != nil {
		return nil, err
	}
	return &matchedImage{URL: publicURL, AssetID: assetID}, nil
}

// pexelsImageWithAsset downloads a Pexels CDN URL, persists the bytes via
// persistImageAsAsset, and returns a matchedImage. The original Pexels
// URL is recorded as the asset's source_url for attribution.
func (s *Server) pexelsImageWithAsset(ctx context.Context, pexelsURL string) (*matchedImage, error) {
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, pexelsURL, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("pexelsImageWithAsset: build request: %w", reqErr)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, getErr := client.Do(req)
	if getErr != nil {
		return nil, fmt.Errorf("pexelsImageWithAsset: download: %w", getErr)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("pexelsImageWithAsset: download status %d", resp.StatusCode)
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
	if readErr != nil {
		return nil, fmt.Errorf("pexelsImageWithAsset: read body: %w", readErr)
	}

	source := pexelsURL
	publicURL, assetID, err := s.persistImageAsAsset(
		ctx,
		"search",
		body,
		asset.SourceKindSearch,
		0,
		&source,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &matchedImage{URL: publicURL, AssetID: assetID}, nil
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

	task, err := s.taskSvc.GetByID(ctx, p.TaskID)
	if err != nil || task == nil {
		s.logger.Warn("image match: task not found, skipping", zap.Int64("task_id", p.TaskID))
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}
	if task.ResultDraftID == nil {
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	mode := task.ImageMode
	if mode == "" {
		mode = taskDomain.ImageModeAuto
	}

	const maxImages = 2
	matched := make([]matchedImage, 0, maxImages)

	// 1. generate_only or auto: try generation first.
	// In generate_only we continue past failures; in auto we break so Pexels can take over.
	if mode == taskDomain.ImageModeGenerateOnly || mode == taskDomain.ImageModeAuto {
		for i := 0; i < maxImages; i++ {
			prompt := fmt.Sprintf("editorial illustration about %s, clean magazine style, high quality", task.Keyword)
			img, genErr := s.generateImageWithAsset(ctx, prompt)
			if genErr != nil {
				s.logger.Warn("image match: generate failed", zap.Error(genErr), zap.Int("idx", i))
				if mode == taskDomain.ImageModeAuto {
					break
				}
				continue
			}
			matched = append(matched, *img)
		}
	}

	// 2. search_only, or auto with insufficient generated, fall back to Pexels.
	needPexels := (mode == taskDomain.ImageModeSearchOnly) ||
		(mode == taskDomain.ImageModeAuto && len(matched) < maxImages)
	if needPexels {
		searchQuery := task.Keyword
		engQuery, engErr := s.callLLM(ctx,
			"Translate the following keyword to a short English search phrase suitable for stock photo search. Return ONLY the English phrase, nothing else.",
			task.Keyword,
			50,
		)
		if engErr == nil && len(engQuery) > 0 && len(engQuery) < 100 {
			searchQuery = engQuery
		}
		imgSvc := pipelinePkg.NewImageService(s.imageSearch, s.imageGen)
		results, searchErr := imgSvc.SearchAndMatch(ctx, searchQuery, maxImages)
		if searchErr != nil {
			s.logger.Warn("image match: search failed", zap.Error(searchErr))
		}
		for _, r := range results {
			if len(matched) >= maxImages {
				break
			}
			img, perr := s.pexelsImageWithAsset(ctx, r.URL)
			if perr != nil {
				s.logger.Warn("image match: pexels persist failed", zap.Error(perr), zap.String("url", r.URL))
				continue
			}
			matched = append(matched, *img)
		}
	}

	if len(matched) == 0 {
		s.logger.Info("image match: no images obtained, continuing")
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	// 3. Embed URLs into draft section blocks AND link asset_id.
	blocks, err := s.blockRepo.FindByDraftID(ctx, *task.ResultDraftID)
	if err != nil || len(blocks) == 0 {
		return s.enqueueNext(pipeline.TypeChartGen, *p)
	}

	imgIdx := 0
	for i, block := range blocks {
		if block.BlockType == "section" && imgIdx < len(matched) {
			imgTag := fmt.Sprintf(`<figure style="margin:20px 0;text-align:center"><img src="%s" alt="%s" style="width:100%%;max-width:100%%;border-radius:12px;display:block" /></figure>`,
				matched[imgIdx].URL, task.Keyword)

			existing := ""
			if block.HTMLFragment != nil {
				existing = *block.HTMLFragment
			}
			newHtml := imgTag + existing
			blocks[i].HTMLFragment = &newHtml
			assetID := matched[imgIdx].AssetID
			blocks[i].AssetID = &assetID
			if err := s.blockRepo.Update(ctx, &blocks[i]); err != nil {
				s.logger.Warn("image match: block update failed",
					zap.Error(err),
					zap.Int64("block_id", blocks[i].ID),
				)
			}

			imgIdx++
			if imgIdx >= maxImages {
				break
			}
		}
	}

	s.logger.Info("image match: assigned images",
		zap.Int("count", imgIdx),
		zap.String("mode", mode),
	)
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

	// Always pre-populate per-block html_fragment using the chosen preset's
	// templates so the frontend preview shows styled HTML even when visual
	// enhance is off. The LLM rewrite below (visual_enhance=true) overwrites
	// these baselines with richer markup; image fragments produced by
	// image_match are preserved.
	styleName := task.ArticleStyle
	if !taskDomain.ValidArticleStyles[styleName] {
		styleName = pipelinePkg.PresetStyleNames()[0]
	}
	baselineSC := pipelinePkg.GetStyleConfig(styleName)
	sectionIdx := 0
	for i := range blocks {
		b := &blocks[i]
		if b.BlockType == draft.BlockTypeSection && b.Heading != nil && strings.TrimSpace(*b.Heading) != "" {
			sectionIdx++
		}
		// Skip blocks that already carry styled HTML (image fragments from image_match).
		if b.HTMLFragment != nil && strings.TrimSpace(*b.HTMLFragment) != "" {
			continue
		}
		styled := pipelinePkg.CompileBlockStyled(*b, baselineSC, sectionIdx)
		if strings.TrimSpace(styled) == "" {
			continue
		}
		b.HTMLFragment = &styled
		if err := s.blockRepo.Update(ctx, b); err != nil {
			s.logger.Warn("baseline html_fragment save failed", zap.Error(err), zap.Int64("block", b.ID))
		}
	}

	// Skip the LLM rewrite if visual enhance is disabled — baselines are enough.
	if !task.VisualEnhance {
		s.logger.Info("visual enhance: baselines applied (LLM skipped)",
			zap.Int64("task_id", p.TaskID),
			zap.String("style", styleName))
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	// Build per-block input for the LLM as a JSON array. Earlier versions
	// concatenated text with byte-level slicing (text[:150]) which:
	//   1. cut Chinese characters mid-byte and produced 还�� garbage,
	//   2. fed only a 150-byte stub per block so the LLM had nothing to
	//      enhance and produced shallow output, and
	//   3. mixed prompt-only annotations like "(已有配图)" into block text,
	//      which the LLM sometimes copied verbatim into the rewritten HTML.
	// Structuring the input keeps annotations in separate fields and
	// preserves the full source text.
	type enhanceInputBlock struct {
		Index    int    `json:"index"`
		Type     string `json:"type"`
		Heading  string `json:"heading,omitempty"`
		Text     string `json:"text,omitempty"`
		HasImage bool   `json:"has_image,omitempty"`
	}
	enhanceInput := make([]enhanceInputBlock, 0, len(blocks))
	for i, b := range blocks {
		enhanceInput = append(enhanceInput, enhanceInputBlock{
			Index:    i,
			Type:     b.BlockType,
			Heading:  derefStr(b.Heading),
			Text:     derefStr(b.TextMD),
			HasImage: b.HTMLFragment != nil && strings.Contains(*b.HTMLFragment, "<img"),
		})
	}
	enhanceInputJSON, err := json.Marshal(enhanceInput)
	if err != nil {
		s.logger.Warn("visual enhance: failed to marshal input", zap.Error(err))
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	// LLM visual enhancement round.
	// The chosen preset (or AI-recommended fallback) maps to a fully specified
	// design system; we feed those tokens into the prompt so the LLM produces
	// inline-styled HTML that matches the reference design.
	enhancePrompt := buildVisualEnhancePrompt(baselineSC)

	enhanceContent, err := s.callLLM(ctx, enhancePrompt, string(enhanceInputJSON), 16384)
	if err != nil {
		s.logger.Warn("visual enhance failed, continuing with original content", zap.Error(err))
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	// Parse enhancement result
	type enhanceBlock struct {
		Index int    `json:"index"`
		HTML  string `json:"html"`
	}
	type enhanceOutput struct {
		StyleName   string          `json:"style_name"`
		ColorScheme json.RawMessage `json:"color_scheme"`
		Blocks      []enhanceBlock  `json:"blocks"`
	}

	var enhanced enhanceOutput
	cleaned := extractJSON(enhanceContent)
	if err := json.Unmarshal([]byte(cleaned), &enhanced); err != nil {
		s.logger.Warn("visual enhance: failed to parse LLM output", zap.Error(err))
		return s.enqueueNext(pipeline.TypeHTMLCompile, *p)
	}

	// Apply enhanced HTML to blocks
	applied := 0
	for _, eb := range enhanced.Blocks {
		if eb.Index >= 0 && eb.Index < len(blocks) && eb.HTML != "" {
			existingHTML := ""
			if blocks[eb.Index].HTMLFragment != nil {
				existingHTML = *blocks[eb.Index].HTMLFragment
			}
			newHTML := eb.HTML
			// Preserve existing images
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

	// Load the draft and compile HTML using the chosen preset.
	task, _ := s.taskSvc.GetByID(ctx, p.TaskID)
	if task != nil && task.ResultDraftID != nil {
		compiler := pipelinePkg.NewCompilerService(s.draftRepo, s.blockRepo)
		if _, err := compiler.CompileStyled(ctx, *task.ResultDraftID, task.ArticleStyle); err != nil {
			s.logger.Warn("html compile failed", zap.Error(err))
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

// buildVisualEnhancePrompt produces a per-preset prompt that gives the LLM the
// exact design-system tokens (palette, fonts, decoration policy) and a few
// gold-standard HTML snippets to imitate per block type.
func buildVisualEnhancePrompt(sc pipelinePkg.StyleConfig) string {
	var refSnippets string
	switch sc.Name {
	case "magazine":
		refSnippets = `参考片段（lead 用 drop-cap）：
<p style="margin:18px 22px;font-size:17px;line-height:1.8;color:#0A0A0A;"><span style="float:left;font-family:'Noto Serif SC',serif;font-weight:900;font-size:60px;line-height:0.9;margin:6px 8px 0 0;color:#E63946;">脑</span>子里有个产品想法，想快速验证一下界面长什么样……</p>

参考片段（H2 + 红色横条）：
<div style="margin:34px 22px 14px;"><div style="width:36px;height:4px;background:#E63946;margin-bottom:12px;"></div><h2 style="margin:0;font-family:'Noto Serif SC',serif;font-weight:900;font-size:26px;line-height:1.2;color:#0A0A0A;">章节标题</h2></div>

参考片段（pull-quote 黑底白字）：
<blockquote style="margin:24px 14px;padding:22px 24px 22px 52px;background:#0A0A0A;color:#F2EFE8;font-family:'Noto Serif SC',serif;font-weight:600;font-size:18px;line-height:1.6;position:relative;"><span style="position:absolute;left:14px;top:-6px;font-family:'Noto Serif SC',serif;font-size:64px;color:#E63946;line-height:1;">&ldquo;</span>金句正文……</blockquote>`
	case "stitch":
		refSnippets = `参考片段（H2 居中 + 短横）：
<h2 style="margin:36px 20px 8px;font-family:'Noto Serif SC',serif;font-weight:700;font-size:22px;color:#D2691E;text-align:center;">章节标题</h2>
<div style="width:56px;height:1px;background:#D2691E;opacity:0.4;margin:0 auto 22px;"></div>

参考片段（强调段橙色加粗）：
<p style="margin:0 20px 22px;color:#D2691E;font-weight:600;font-size:16px;line-height:1.85;">这一段需要醒目，用橙色加粗。</p>

参考片段（输入引用）：
<div style="margin:22px 12px;padding:18px 20px;background:#FDF2E5;border-left:3px solid #D2691E;border-radius:6px;font-size:14px;line-height:1.85;color:#1A1815;">输入示例……</div>`
	default: // minimal
		refSnippets = `参考片段（H2 + mono 编号）：
<h2 style="margin:36px 24px 14px;font-family:'Noto Serif SC',serif;font-weight:700;font-size:20px;line-height:1.4;color:#111111;display:flex;align-items:baseline;gap:12px;"><span style="font-family:'JetBrains Mono',monospace;font-size:11px;color:#888888;letter-spacing:1px;">01</span><span>章节标题</span></h2>

参考片段（黄色高亮 span）：
<span style="background:linear-gradient(transparent 60%,#FFE94D 60%);padding:0 2px;">需要被强调的短语</span>

参考片段（行内 code）：
<code style="background:#FAFAFA;color:#111111;padding:1px 6px;border-radius:3px;font-family:'JetBrains Mono',monospace;font-size:13px;">--global</code>`
	}

	return fmt.Sprintf(`你是 ReadBud 公众号排版工程师。本次请把每个 block 重排成符合下面预设设计系统的 inline-styled HTML。

# 预设：%s（%s）
%s

## 输入格式
用户消息是一个严格的 JSON 数组，每一项形如：
{"index": <int>, "type": "<block类型>", "heading": "<可选>", "text": "<完整正文，可能有换行>", "has_image": <bool>}
- text 字段是完整的正文（不会被截断）。请基于完整 text 重排版面，不要复读、缩短或省略。
- has_image=true 表示该 block 已经挂了一张图（在 ReadBud 内部 html_fragment 中），你只负责文本部分的样式，**不要**在你的输出 HTML 里写"已有配图"、"(已有配图)"、"with image"、"index"、"has_image"、"type" 这些元字段或注释。
- heading 字段如果存在，请按当前预设的 H2 装饰策略渲染。

## 设计 token（必须使用）
- 主底色 paper: %s
- 面板底色 paperAlt: %s
- 主文字色 ink: %s
- 正文色 body: %s
- 次要色 mute: %s
- 分隔线 line: %s
- 强调色 accent: %s
- 强调底色 accentSoft: %s
- 衬线字族 serif: %s
- 无衬线字族 sans: %s
- 等宽字族 mono: %s
- H2 装饰策略：%s（minimal=mono编号 / magazine=红色横条 / stitch=居中短横）
- 导语装饰策略：%s（lead-text=底部细线 / drop-cap=红色首字下沉 / box=橙色底块）

## 排版约束
- 正文字号 15-17px，行高 1.75-1.85
- 段落两侧留 22-24px 内边距（margin-left/right），不要顶到屏幕边
- 配色严格使用上面 8 个 token，禁止引入其它颜色
- 强调只用 1 种方式：%s 用 ==黄色高亮 span==；%s 用 红色 italic em；%s 用 橙色 bold
- **必须保留 text 字段的全部文字**，可以拆段、加强调，但不能删字、不能省略段落、不能用"..."代替
- 已有的 <img>、<figure>、<svg> 标签必须原样保留
- 行内 code 保留为 <code>，使用上面的样式

## 严禁
- 在输出 HTML 中出现输入 JSON 的字段名或值（例如 "(已有配图)"、"index=0"、"has_image"）
- class 选择器、id 选择器
- @keyframes、CSS 动画
- position:absolute / fixed
- JavaScript、<script>
- linear-gradient 之外的 gradient
- 任意外部字体或图片资源

## 参考片段
%s

## 输出格式（严格 JSON，不要 markdown 代码块）
{
  "style_name": %q,
  "blocks": [
    {"index": 0, "html": "对应 block 的完整 HTML，含 inline style；保留 text 全文"},
    {"index": 1, "html": "..."}
  ]
}`,
		sc.Name, sc.DisplayName, sc.Description,
		sc.Paper, sc.PaperAlt, sc.Ink, sc.Body, sc.Mute, sc.Line, sc.Accent, sc.AccentSoft,
		sc.SerifStack, sc.SansStack, sc.MonoStack,
		sc.H2Decor, sc.LeadDecor,
		"minimal", "magazine", "stitch",
		refSnippets,
		sc.Name)
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
