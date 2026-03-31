package service

import (
	"context"
	"fmt"

	"readbud/internal/api/dto"
	taskDomain "readbud/internal/domain/task"
	"readbud/internal/pkg/sse"
	"readbud/internal/pkg/utils"
	"readbud/internal/repository/postgres"
)

// TaskService handles content task business logic.
type TaskService struct {
	taskRepo postgres.TaskRepository
	sseHub   *sse.Hub
}

// NewTaskService creates a new TaskService.
func NewTaskService(taskRepo postgres.TaskRepository, sseHub *sse.Hub) *TaskService {
	return &TaskService{taskRepo: taskRepo, sseHub: sseHub}
}

// Create creates a new content task and returns the view object.
func (s *TaskService) Create(ctx context.Context, req dto.CreateTaskRequest) (*dto.TaskVO, error) {
	t := taskDomain.ContentTask{
		TaskNo:      utils.NewULID(),
		Keyword:     req.Keyword,
		Audience:    req.Audience,
		Tone:        req.Tone,
		TargetWords: req.TargetWords,
		ImageMode:   req.ImageMode,
		ChartMode:   req.ChartMode,
		PublishMode: req.PublishMode,
		PublishAt:   req.PublishAt,
		Status:      taskDomain.StatusPending,
		Progress:    0,
	}

	if t.TargetWords == 0 {
		t.TargetWords = 2000
	}

	if err := s.taskRepo.Create(ctx, &t); err != nil {
		return nil, fmt.Errorf("taskService.Create: %w", err)
	}

	vo := taskToVO(t)
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
	vo := taskToVO(*t)
	return &vo, nil
}

// ListRecent returns a paginated list of recent tasks.
func (s *TaskService) ListRecent(ctx context.Context, page, pageSize int) (*dto.TaskListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	tasks, total, err := s.taskRepo.ListRecent(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("taskService.ListRecent: %w", err)
	}

	items := make([]dto.TaskVO, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, taskToVO(t))
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
	s.sseHub.Publish(t.PublicID, sse.Event{
		Type: "progress",
		Data: map[string]interface{}{
			"status":   status,
			"stage":    stage,
			"progress": progress,
		},
	})
	return nil
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
	s.sseHub.Publish(t.PublicID, sse.Event{
		Type: "failed",
		Data: map[string]interface{}{
			"status":  taskDomain.StatusFailed,
			"message": errMsg,
		},
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

	vo := taskToVO(*t)
	return &vo, nil
}

func taskToVO(t taskDomain.ContentTask) dto.TaskVO {
	return dto.TaskVO{
		ID:           t.PublicID,
		TaskNo:       t.TaskNo,
		Keyword:      t.Keyword,
		Audience:     t.Audience,
		Tone:         t.Tone,
		TargetWords:  t.TargetWords,
		ImageMode:    t.ImageMode,
		ChartMode:    t.ChartMode,
		PublishMode:  t.PublishMode,
		PublishAt:    t.PublishAt,
		Status:       t.Status,
		Progress:     t.Progress,
		CurrentStage: t.CurrentStage,
		ErrorMessage: t.ErrorMessage,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
