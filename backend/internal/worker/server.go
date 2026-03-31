package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	taskDomain "readbud/internal/domain/task"
	"readbud/internal/service"
)

// Server wraps the Asynq server and registers pipeline handlers.
type Server struct {
	srv     *asynq.Server
	mux     *asynq.ServeMux
	taskSvc *service.TaskService
}

// ServerConfig holds configuration for the Asynq worker server.
type ServerConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
}

// NewServer creates a new Asynq worker server.
func NewServer(cfg ServerConfig, taskSvc *service.TaskService) *Server {
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

	return &Server{
		srv:     srv,
		mux:     asynq.NewServeMux(),
		taskSvc: taskSvc,
	}
}

// RegisterHandlers registers all pipeline stage handlers.
func (s *Server) RegisterHandlers() {
	s.mux.HandleFunc(TypeKeywordExpand, s.handleKeywordExpand)
	s.mux.HandleFunc(TypeSourceSearch, s.handleSourceSearch)
	s.mux.HandleFunc(TypeContentCrawl, s.handleContentCrawl)
	s.mux.HandleFunc(TypeHotScore, s.handleHotScore)
	s.mux.HandleFunc(TypeArticleWrite, s.handleArticleWrite)
	s.mux.HandleFunc(TypeImageMatch, s.handleImageMatch)
	s.mux.HandleFunc(TypeChartGen, s.handleChartGen)
	s.mux.HandleFunc(TypeHTMLCompile, s.handleHTMLCompile)
	s.mux.HandleFunc(TypePublish, s.handlePublish)
}

// Start starts the Asynq server.
func (s *Server) Start() error {
	s.RegisterHandlers()
	return s.srv.Start(s.mux)
}

// Shutdown gracefully stops the Asynq server.
func (s *Server) Shutdown() {
	s.srv.Shutdown()
}

// ---------- Pipeline Stage Handlers ----------
// Each handler processes a pipeline stage and enqueues the next stage on success.
// Stub implementations for now — real logic will be added in subsequent phases.

func (s *Server) handleKeywordExpand(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageKeywordExpand, 10); err != nil {
		return fmt.Errorf("handleKeywordExpand: update progress: %w", err)
	}

	// TODO: Call LLMProvider to expand keyword into search queries
	log.Printf("[pipeline] keyword expand for task %d (stub)", p.TaskID)

	return s.enqueueNext(TypeSourceSearch, *p)
}

func (s *Server) handleSourceSearch(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageSourceSearch, 20); err != nil {
		return fmt.Errorf("handleSourceSearch: update progress: %w", err)
	}

	// TODO: Call SearchProvider to search for source articles
	log.Printf("[pipeline] source search for task %d (stub)", p.TaskID)

	return s.enqueueNext(TypeContentCrawl, *p)
}

func (s *Server) handleContentCrawl(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageContentCrawl, 35); err != nil {
		return fmt.Errorf("handleContentCrawl: update progress: %w", err)
	}

	// TODO: Call CrawlerProvider to crawl source URLs
	log.Printf("[pipeline] content crawl for task %d (stub)", p.TaskID)

	return s.enqueueNext(TypeHotScore, *p)
}

func (s *Server) handleHotScore(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageHotScore, 45); err != nil {
		return fmt.Errorf("handleHotScore: update progress: %w", err)
	}

	// TODO: Calculate hot scores for crawled content
	log.Printf("[pipeline] hot score for task %d (stub)", p.TaskID)

	return s.enqueueNext(TypeArticleWrite, *p)
}

func (s *Server) handleArticleWrite(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageArticleWrite, 60); err != nil {
		return fmt.Errorf("handleArticleWrite: update progress: %w", err)
	}

	// TODO: Call LLMProvider to generate article draft
	log.Printf("[pipeline] article write for task %d (stub)", p.TaskID)

	return s.enqueueNext(TypeImageMatch, *p)
}

func (s *Server) handleImageMatch(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageImageMatch, 75); err != nil {
		return fmt.Errorf("handleImageMatch: update progress: %w", err)
	}

	// TODO: Search + generate images for article
	log.Printf("[pipeline] image match for task %d (stub)", p.TaskID)

	return s.enqueueNext(TypeChartGen, *p)
}

func (s *Server) handleChartGen(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageChartGen, 85); err != nil {
		return fmt.Errorf("handleChartGen: update progress: %w", err)
	}

	// TODO: Generate charts from extracted data
	log.Printf("[pipeline] chart gen for task %d (stub)", p.TaskID)

	return s.enqueueNext(TypeHTMLCompile, *p)
}

func (s *Server) handleHTMLCompile(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StageHTMLCompile, 95); err != nil {
		return fmt.Errorf("handleHTMLCompile: update progress: %w", err)
	}

	// TODO: Compile article blocks + images + charts into WeChat HTML
	log.Printf("[pipeline] HTML compile for task %d (stub)", p.TaskID)

	// Final stage depends on publish mode — for now mark as done
	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusDone, taskDomain.StageHTMLCompile, 100); err != nil {
		return fmt.Errorf("handleHTMLCompile: mark done: %w", err)
	}
	return nil
}

func (s *Server) handlePublish(ctx context.Context, t *asynq.Task) error {
	p, err := ParsePipelinePayload(t)
	if err != nil {
		return err
	}

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusRunning, taskDomain.StagePublish, 98); err != nil {
		return fmt.Errorf("handlePublish: update progress: %w", err)
	}

	// TODO: Call WeChatPublisher to publish the compiled article
	log.Printf("[pipeline] publish for task %d (stub)", p.TaskID)

	if err := s.taskSvc.UpdateProgress(ctx, p.TaskID, taskDomain.StatusDone, taskDomain.StagePublish, 100); err != nil {
		return fmt.Errorf("handlePublish: mark done: %w", err)
	}
	return nil
}

// enqueueNext is a helper — in production this uses an injected asynq.Client.
// For now it's a placeholder that will be replaced when the client is wired in.
func (s *Server) enqueueNext(taskType string, payload PipelinePayload) error {
	// TODO: Use injected asynq.Client to enqueue next stage
	log.Printf("[pipeline] would enqueue %s for task %d", taskType, payload.TaskID)
	return nil
}
