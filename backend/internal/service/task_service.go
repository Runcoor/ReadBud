package service

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"

	"readbud/internal/api/dto"
	taskDomain "readbud/internal/domain/task"
	"readbud/internal/pkg/sse"
	"readbud/internal/pkg/utils"
	"readbud/internal/repository/postgres"
	"readbud/internal/pipeline"
)

// TaskService handles content task business logic.
type TaskService struct {
	taskRepo    postgres.TaskRepository
	draftRepo   postgres.ArticleDraftRepository
	publisher   sse.EventPublisher
	asynqClient *asynq.Client
	brandRepo   postgres.BrandProfileRepository
}

// NewTaskService creates a new TaskService.
func NewTaskService(taskRepo postgres.TaskRepository, draftRepo postgres.ArticleDraftRepository, publisher sse.EventPublisher, asynqClient *asynq.Client, brandRepo postgres.BrandProfileRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo, draftRepo: draftRepo, publisher: publisher, asynqClient: asynqClient, brandRepo: brandRepo}
}

// Create creates a new content task and enqueues the pipeline.
func (s *TaskService) Create(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskVO, error) {
	t := taskDomain.ContentTask{
		TaskNo:      utils.NewULID(),
		Keyword:     req.Keyword,
		Audience:    req.Audience,
		Tone:        req.Tone,
		TargetWords: req.TargetWords,
		ImageMode:   req.ImageMode,
		ChartMode:   req.ChartMode,
		PublishMode:   req.PublishMode,
		PublishAt:     req.PublishAt,
		ArticleStyle: req.ArticleStyle,
		Status:        taskDomain.StatusPending,
		Progress:    0,
	}

	if req.VisualEnhance != nil {
		t.VisualEnhance = *req.VisualEnhance
	}

	if req.BrandProfileID != nil && *req.BrandProfileID != "" {
		bp, bpErr := s.brandRepo.FindByPublicID(ctx, *req.BrandProfileID)
		if bpErr == nil && bp != nil {
			t.BrandProfileID = &bp.ID
		}
	}

	if t.TargetWords == 0 {
		t.TargetWords = 2000
	}

	if err := s.taskRepo.Create(ctx, &t); err != nil {
		return nil, fmt.Errorf("taskService.Create: %w", err)
	}

	// Enqueue the first pipeline stage
	if s.asynqClient != nil {
		payload := pipeline.Payload{
			TaskID:   t.ID,
			PublicID: t.PublicID,
		}
		task, err := pipeline.NewTask(pipeline.TypeKeywordExpand, payload)
		if err != nil {
			log.Printf("[task] failed to create pipeline task: %v", err)
		} else if _, err := s.asynqClient.Enqueue(task); err != nil {
			log.Printf("[task] failed to enqueue pipeline task: %v", err)
		}
	}

	vo := taskToVO(t, nil, nil)
	return &vo, nil
}

// GetByPublicID returns a task by its public ID.
func (s *TaskService) GetByPublicID(ctx context.Context, publicID string) (*dto.TaskVO, error) {
	t, err := s.taskRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("taskService.GetByPublicID: %w", err)
	}
	if t == nil {
		return nil, ErrNotFound
	}
	vo := taskToVO(*t, s.resolveDraftPublicID(ctx, t.ResultDraftID), s.resolveBrandPublicID(ctx, t.BrandProfileID))
	return &vo, nil
}

// ListRecent returns a paginated list of recent tasks, optionally filtered by status.
func (s *TaskService) ListRecent(ctx context.Context, page, pageSize int, status string) (*dto.TaskListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var tasks []taskDomain.ContentTask
	var total int64
	var err error

	if status != "" {
		tasks, total, err = s.taskRepo.ListByStatus(ctx, status, pageSize, offset)
	} else {
		tasks, total, err = s.taskRepo.ListRecent(ctx, pageSize, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("taskService.ListRecent: %w", err)
	}

	items := make([]dto.TaskVO, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, taskToVO(t, s.resolveDraftPublicID(ctx, t.ResultDraftID), s.resolveBrandPublicID(ctx, t.BrandProfileID)))
	}

	return &dto.TaskListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateProgress updates a task's progress and current stage.
func (s *TaskService) UpdateProgress(ctx context.Context, taskID int64, status string, stage string, progress int) error {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("taskService.UpdateProgress: %w", err)
	}
	if t == nil {
		return ErrNotFound
	}

	t.Status = status
	t.CurrentStage = stage
	t.Progress = progress

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return fmt.Errorf("taskService.UpdateProgress: %w", err)
	}

	// Publish SSE event for real-time progress
	s.publisher.Publish(t.PublicID, sse.Event{
		Type: "progress",
		Data: map[string]interface{}{
			"status":   status,
			"stage":    stage,
			"progress": progress,
		},
	})
	return nil
}

// SetResultDraft links a draft to the task.
func (s *TaskService) SetResultDraft(ctx context.Context, taskID int64, draftID int64) error {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("taskService.SetResultDraft: %w", err)
	}
	if t == nil {
		return ErrNotFound
	}
	t.ResultDraftID = &draftID
	return s.taskRepo.Update(ctx, t)
}

// MarkFailed marks a task as failed with an error message.
func (s *TaskService) MarkFailed(ctx context.Context, taskID int64, errMsg string) error {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("taskService.MarkFailed: %w", err)
	}
	if t == nil {
		return ErrNotFound
	}

	t.Status = taskDomain.StatusFailed
	t.ErrorMessage = &errMsg

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return fmt.Errorf("taskService.MarkFailed: %w", err)
	}

	// Publish SSE event for failure
	s.publisher.Publish(t.PublicID, sse.Event{
		Type: "failed",
		Data: map[string]interface{}{
			"status":  taskDomain.StatusFailed,
			"message": errMsg,
		},
	})
	return nil
}

// MarkDone marks a task as completed.
func (s *TaskService) MarkDone(ctx context.Context, taskID int64) error {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("taskService.MarkDone: %w", err)
	}
	if t == nil {
		return ErrNotFound
	}

	t.Status = taskDomain.StatusDone
	t.Progress = 100

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return fmt.Errorf("taskService.MarkDone: %w", err)
	}

	doneData := map[string]interface{}{
		"status":   taskDomain.StatusDone,
		"progress": 100,
	}
	// Include draft public ID so frontend can show preview immediately
	if draftPubID := s.resolveDraftPublicID(ctx, t.ResultDraftID); draftPubID != nil {
		doneData["result_draft_id"] = *draftPubID
	}
	s.publisher.Publish(t.PublicID, sse.Event{
		Type: "done",
		Data: doneData,
	})
	return nil
}

// Retry resets a failed task for re-execution.
func (s *TaskService) Retry(ctx context.Context, publicID string) (*dto.TaskVO, error) {
	t, err := s.taskRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("taskService.Retry: %w", err)
	}
	if t == nil {
		return nil, ErrNotFound
	}
	if t.Status != taskDomain.StatusFailed {
		return nil, fmt.Errorf("taskService.Retry: task status is %s, expected failed", t.Status)
	}

	t.Status = taskDomain.StatusPending
	t.ErrorMessage = nil
	t.Progress = 0
	t.CurrentStage = ""

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("taskService.Retry: %w", err)
	}

	// Re-enqueue pipeline
	if s.asynqClient != nil {
		payload := pipeline.Payload{TaskID: t.ID, PublicID: t.PublicID}
		task, _ := pipeline.NewTask(pipeline.TypeKeywordExpand, payload)
		if task != nil {
			s.asynqClient.Enqueue(task)
		}
	}

	vo := taskToVO(*t, nil, s.resolveBrandPublicID(ctx, t.BrandProfileID))
	return &vo, nil
}

// CancelTask cancels a pending or running task.
func (s *TaskService) CancelTask(ctx context.Context, publicID string) error {
	t, err := s.taskRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("taskService.CancelTask: %w", err)
	}
	if t == nil {
		return ErrNotFound
	}
	if t.Status != taskDomain.StatusPending && t.Status != taskDomain.StatusRunning {
		return ErrInvalidState
	}

	t.Status = taskDomain.StatusCancelled
	if err := s.taskRepo.Update(ctx, t); err != nil {
		return fmt.Errorf("taskService.CancelTask: %w", err)
	}

	s.publisher.Publish(t.PublicID, sse.Event{
		Type: "cancelled",
		Data: map[string]interface{}{
			"status": taskDomain.StatusCancelled,
		},
	})
	return nil
}

// UpdateArticleStyle updates the article style for a task.
func (s *TaskService) UpdateArticleStyle(ctx context.Context, taskID int64, style string) error {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil || t == nil {
		return fmt.Errorf("taskService.UpdateArticleStyle: task not found")
	}
	t.ArticleStyle = style
	return s.taskRepo.Update(ctx, t)
}

// GetByID returns a task by its internal ID.
func (s *TaskService) GetByID(ctx context.Context, id int64) (*taskDomain.ContentTask, error) {
	return s.taskRepo.FindByID(ctx, id)
}

func (s *TaskService) resolveBrandPublicID(ctx context.Context, brandID *int64) *string {
	if brandID == nil || s.brandRepo == nil {
		return nil
	}
	bp, err := s.brandRepo.FindByID(ctx, *brandID)
	if err != nil || bp == nil {
		return nil
	}
	return &bp.PublicID
}

func (s *TaskService) resolveDraftPublicID(ctx context.Context, draftID *int64) *string {
	if draftID == nil || s.draftRepo == nil {
		return nil
	}
	d, err := s.draftRepo.FindByID(ctx, *draftID)
	if err != nil || d == nil {
		return nil
	}
	return &d.PublicID
}

func taskToVO(t taskDomain.ContentTask, draftPublicID *string, brandPublicID *string) dto.TaskVO {
	return dto.TaskVO{
		ID:             t.PublicID,
		TaskNo:         t.TaskNo,
		Keyword:        t.Keyword,
		Audience:       t.Audience,
		Tone:           t.Tone,
		TargetWords:    t.TargetWords,
		ImageMode:      t.ImageMode,
		ChartMode:      t.ChartMode,
		PublishMode:    t.PublishMode,
		PublishAt:      t.PublishAt,
		Status:         t.Status,
		Progress:       t.Progress,
		CurrentStage:   t.CurrentStage,
		ErrorMessage:   t.ErrorMessage,
		ResultDraftID:  draftPublicID,
		ArticleStyle:   t.ArticleStyle,
		VisualEnhance:  t.VisualEnhance,
		BrandProfileID: brandPublicID,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
