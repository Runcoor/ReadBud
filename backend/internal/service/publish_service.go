package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain/publish"
	"readbud/internal/integration/wechat"
	"readbud/internal/repository/postgres"
)

// PublishService handles publish job orchestration.
type PublishService struct {
	jobRepo         postgres.PublishJobRepository
	recordRepo      postgres.PublishRecordRepository
	publisher       adapter.WeChatPublisher
	tokenProv       wechat.TokenProvider
	contentImageSvc *ContentImageService
	logger          *zap.Logger
}

// NewPublishService creates a new PublishService.
func NewPublishService(
	jobRepo postgres.PublishJobRepository,
	recordRepo postgres.PublishRecordRepository,
	publisher adapter.WeChatPublisher,
	tokenProv wechat.TokenProvider,
	contentImageSvc *ContentImageService,
	logger *zap.Logger,
) *PublishService {
	return &PublishService{
		jobRepo:         jobRepo,
		recordRepo:      recordRepo,
		publisher:       publisher,
		tokenProv:       tokenProv,
		contentImageSvc: contentImageSvc,
		logger:          logger,
	}
}

// CreateJob creates a new publish job in queued status.
func (s *PublishService) CreateJob(ctx context.Context, draftID, wechatAccountID int64, publishMode string) (*publish.PublishJob, error) {
	job := &publish.PublishJob{
		DraftID:         draftID,
		WechatAccountID: wechatAccountID,
		PublishMode:     publishMode,
		Status:          publish.JobStatusQueued,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("publishService.CreateJob: %w", err)
	}

	s.logger.Info("publish job created",
		zap.Int64("job_id", job.ID),
		zap.Int64("draft_id", draftID),
		zap.String("mode", publishMode),
	)
	return job, nil
}

// GetJob retrieves a publish job by its public ID.
func (s *PublishService) GetJob(ctx context.Context, publicID string) (*publish.PublishJob, error) {
	job, err := s.jobRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("publishService.GetJob: %w", err)
	}
	if job == nil {
		return nil, ErrNotFound
	}
	return job, nil
}

// CancelJob cancels a publish job if it is in a cancellable state.
func (s *PublishService) CancelJob(ctx context.Context, publicID string) (*publish.PublishJob, error) {
	job, err := s.jobRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("publishService.CancelJob: %w", err)
	}
	if job == nil {
		return nil, ErrNotFound
	}

	switch job.Status {
	case publish.JobStatusQueued, publish.JobStatusSubmitting, publish.JobStatusPolling:
		// Cancellable states
	default:
		return nil, fmt.Errorf("publishService.CancelJob: job status %s is not cancellable", job.Status)
	}

	job.Status = publish.JobStatusCancelled
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("publishService.CancelJob: %w", err)
	}

	s.logger.Info("publish job cancelled", zap.String("public_id", publicID))
	return job, nil
}

// ProcessJob executes the stub publish pipeline: create draft -> submit -> poll.
// Status flow: queued -> submitting -> polling -> success/failed.
func (s *PublishService) ProcessJob(ctx context.Context, jobID int64, appID string) error {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("publishService.ProcessJob: %w", err)
	}
	if job == nil {
		return ErrNotFound
	}

	if job.Status == publish.JobStatusCancelled {
		return nil
	}

	// Step 1: submitting — create draft on WeChat
	job.Status = publish.JobStatusSubmitting
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("publishService.ProcessJob: update to submitting: %w", err)
	}

	token, err := s.tokenProv.GetAccessToken(ctx, appID)
	if err != nil {
		return s.failJob(ctx, job, fmt.Sprintf("get access token: %v", err))
	}

	// Upload content images to WeChat before HTML compilation.
	// WeChat filters external URLs, so all images must go through their API.
	if s.contentImageSvc != nil {
		uploadResults, err := s.contentImageSvc.UploadForDraft(ctx, job.DraftID, appID)
		if err != nil {
			s.logger.Warn("content image upload had errors, proceeding with available images",
				zap.Int64("draft_id", job.DraftID),
				zap.Error(err),
			)
		}
		if len(uploadResults) > 0 {
			s.logger.Info("content images uploaded for draft",
				zap.Int64("draft_id", job.DraftID),
				zap.Int("count", len(uploadResults)),
			)
		}
	}

	article := adapter.WeChatArticle{
		Title:   fmt.Sprintf("Draft %d", job.DraftID),
		Content: "<p>Stub content for publish pipeline</p>",
	}

	// Replace image URLs in HTML with WeChat URLs
	if s.contentImageSvc != nil {
		replaced, err := s.contentImageSvc.ReplaceImageURLsInHTML(ctx, article.Content, job.DraftID)
		if err != nil {
			s.logger.Warn("failed to replace image URLs in HTML",
				zap.Int64("draft_id", job.DraftID),
				zap.Error(err),
			)
		} else {
			article.Content = replaced
		}
	}

	mediaID, err := s.publisher.CreateDraft(ctx, token, article)
	if err != nil {
		return s.failJob(ctx, job, fmt.Sprintf("create draft: %v", err))
	}

	// Step 2: polling — submit for publish
	job.Status = publish.JobStatusPolling
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("publishService.ProcessJob: update to polling: %w", err)
	}

	result, err := s.publisher.Publish(ctx, token, mediaID)
	if err != nil {
		return s.failJob(ctx, job, fmt.Sprintf("publish: %v", err))
	}

	// Step 3: success — create publish record
	record := &publish.PublishRecord{
		PublishJobID:    job.ID,
		DraftID:         job.DraftID,
		WechatAccountID: job.WechatAccountID,
		DraftMediaID:    &mediaID,
		PublishID:       &result.PublishID,
		PublishStatus:   publish.JobStatusSuccess,
	}
	if err := s.recordRepo.Create(ctx, record); err != nil {
		return fmt.Errorf("publishService.ProcessJob: create record: %w", err)
	}

	job.Status = publish.JobStatusSuccess
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("publishService.ProcessJob: update to success: %w", err)
	}

	s.logger.Info("publish job completed",
		zap.Int64("job_id", job.ID),
		zap.String("publish_id", result.PublishID),
	)
	return nil
}

// failJob marks a job as failed with the given error message.
func (s *PublishService) failJob(ctx context.Context, job *publish.PublishJob, errMsg string) error {
	job.Status = publish.JobStatusFailed
	job.LastError = &errMsg
	job.RetryCount++

	if err := s.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("publishService.failJob: %w", err)
	}

	s.logger.Error("publish job failed",
		zap.Int64("job_id", job.ID),
		zap.String("error", errMsg),
	)
	return fmt.Errorf("publishService.ProcessJob: %s", errMsg)
}
