package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain/task"
)

// TaskRepository defines the interface for task data access.
type TaskRepository interface {
	Create(ctx context.Context, t *task.ContentTask) error
	FindByID(ctx context.Context, id int64) (*task.ContentTask, error)
	FindByPublicID(ctx context.Context, publicID string) (*task.ContentTask, error)
	FindByTaskNo(ctx context.Context, taskNo string) (*task.ContentTask, error)
	Update(ctx context.Context, t *task.ContentTask) error
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]task.ContentTask, int64, error)
	ListRecent(ctx context.Context, limit, offset int) ([]task.ContentTask, int64, error)
}

type taskRepo struct {
	db *gorm.DB
}

// NewTaskRepository creates a new PostgreSQL-backed task repository.
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepo{db: db}
}

func (r *taskRepo) Create(ctx context.Context, t *task.ContentTask) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("taskRepo.Create: %w", err)
	}
	return nil
}

func (r *taskRepo) FindByID(ctx context.Context, id int64) (*task.ContentTask, error) {
	var t task.ContentTask
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("taskRepo.FindByID: %w", err)
	}
	return &t, nil
}

func (r *taskRepo) FindByPublicID(ctx context.Context, publicID string) (*task.ContentTask, error) {
	var t task.ContentTask
	if err := r.db.WithContext(ctx).Where("public_id = ?", publicID).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("taskRepo.FindByPublicID: %w", err)
	}
	return &t, nil
}

func (r *taskRepo) FindByTaskNo(ctx context.Context, taskNo string) (*task.ContentTask, error) {
	var t task.ContentTask
	if err := r.db.WithContext(ctx).Where("task_no = ?", taskNo).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("taskRepo.FindByTaskNo: %w", err)
	}
	return &t, nil
}

func (r *taskRepo) Update(ctx context.Context, t *task.ContentTask) error {
	if err := r.db.WithContext(ctx).Save(t).Error; err != nil {
		return fmt.Errorf("taskRepo.Update: %w", err)
	}
	return nil
}

func (r *taskRepo) ListByStatus(ctx context.Context, status string, limit, offset int) ([]task.ContentTask, int64, error) {
	var tasks []task.ContentTask
	var total int64

	q := r.db.WithContext(ctx).Model(&task.ContentTask{}).Where("status = ?", status)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("taskRepo.ListByStatus count: %w", err)
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("taskRepo.ListByStatus: %w", err)
	}
	return tasks, total, nil
}

func (r *taskRepo) ListRecent(ctx context.Context, limit, offset int) ([]task.ContentTask, int64, error) {
	var tasks []task.ContentTask
	var total int64

	q := r.db.WithContext(ctx).Model(&task.ContentTask{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("taskRepo.ListRecent count: %w", err)
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("taskRepo.ListRecent: %w", err)
	}
	return tasks, total, nil
}
