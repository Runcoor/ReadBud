// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
	"readbud/internal/domain"
	"readbud/internal/domain/publish"
	"readbud/internal/integration/wechat"
	"readbud/internal/repository/postgres"
)

// PublishService handles publish job orchestration.
type PublishService struct {
	jobRepo         postgres.PublishJobRepository
	recordRepo      postgres.PublishRecordRepository
	draftRepo       postgres.ArticleDraftRepository
	wechatRepo      postgres.WechatAccountRepository
	publisher       adapter.WeChatPublisher
	tokenProv       wechat.TokenProvider
	contentImageSvc *ContentImageService
	coverImageSvc   *CoverImageService
	logger          *zap.Logger
}

// NewPublishService creates a new PublishService. coverImageSvc may be nil; if so,
// a missing draft cover will surface as an error from CreateDraft (WeChat requires
// thumb_media_id) instead of being auto-generated.
func NewPublishService(
	jobRepo postgres.PublishJobRepository,
	recordRepo postgres.PublishRecordRepository,
	draftRepo postgres.ArticleDraftRepository,
	wechatRepo postgres.WechatAccountRepository,
	publisher adapter.WeChatPublisher,
	tokenProv wechat.TokenProvider,
	contentImageSvc *ContentImageService,
	coverImageSvc *CoverImageService,
	logger *zap.Logger,
) *PublishService {
	return &PublishService{
		jobRepo:         jobRepo,
		recordRepo:      recordRepo,
		draftRepo:       draftRepo,
		wechatRepo:      wechatRepo,
		publisher:       publisher,
		tokenProv:       tokenProv,
		contentImageSvc: contentImageSvc,
		coverImageSvc:   coverImageSvc,
		logger:          logger,
	}
}

// CreateJob creates a new publish job. The initial status depends on the
// WechatAccount.delivery_mode of the target account:
//
//	api       -> queued              (will be picked up by the worker for API publish)
//	extension -> awaiting_extension  (browser plugin will fulfill it)
//	manual    -> awaiting_manual     (user copies into editor by hand)
//
// The publish_mode argument here is the orthogonal "now/schedule/manual" timing
// concept and is stored as-is on the job. Delivery mode is read off the account.
func (s *PublishService) CreateJob(ctx context.Context, draftID, wechatAccountID int64, publishMode string) (*publish.PublishJob, error) {
	delivery := domain.DeliveryModeAPI
	if s.wechatRepo != nil {
		acct, err := s.wechatRepo.FindByID(ctx, wechatAccountID)
		if err != nil {
			return nil, fmt.Errorf("publishService.CreateJob: load account: %w", err)
		}
		if acct != nil && acct.DeliveryMode != "" {
			delivery = acct.DeliveryMode
		}
	}

	job := &publish.PublishJob{
		DraftID:         draftID,
		WechatAccountID: wechatAccountID,
		PublishMode:     publishMode,
		Status:          initialStatusForDelivery(delivery),
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("publishService.CreateJob: %w", err)
	}

	s.logger.Info("publish job created",
		zap.Int64("job_id", job.ID),
		zap.Int64("draft_id", draftID),
		zap.String("mode", publishMode),
		zap.String("delivery", delivery),
		zap.String("status", job.Status),
	)
	return job, nil
}

// initialStatusForDelivery maps a delivery mode to the publish-job status it
// should be created in. Defaulting to the queued/api path keeps existing
// behavior intact when the account has no explicit delivery_mode set.
func initialStatusForDelivery(delivery string) string {
	switch delivery {
	case domain.DeliveryModeExtension:
		return publish.JobStatusAwaitingExtension
	case domain.DeliveryModeManual:
		return publish.JobStatusAwaitingManual
	default:
		return publish.JobStatusQueued
	}
}

// MarkExtensionFulfilled records that the browser extension has finished filling
// (and the user has hit "群发" inside the WeChat editor). Used by the plugin
// callback endpoint to flip awaiting_extension -> success without going through
// the API-publish state machine. ArticleURL is optional but recommended.
func (s *PublishService) MarkExtensionFulfilled(ctx context.Context, publicID, articleURL string) (*publish.PublishJob, error) {
	job, err := s.jobRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("publishService.MarkExtensionFulfilled: %w", err)
	}
	if job == nil {
		return nil, ErrNotFound
	}

	if job.Status != publish.JobStatusAwaitingExtension && job.Status != publish.JobStatusAwaitingManual {
		return nil, fmt.Errorf("publishService.MarkExtensionFulfilled: job in status %s cannot be marked fulfilled", job.Status)
	}

	record := &publish.PublishRecord{
		PublishJobID:    job.ID,
		DraftID:         job.DraftID,
		WechatAccountID: job.WechatAccountID,
		PublishStatus:   publish.JobStatusSuccess,
	}
	if articleURL != "" {
		url := articleURL
		record.ArticleURL = &url
	}
	if err := s.recordRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("publishService.MarkExtensionFulfilled: create record: %w", err)
	}

	job.Status = publish.JobStatusSuccess
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("publishService.MarkExtensionFulfilled: update job: %w", err)
	}
	s.logger.Info("publish job fulfilled via extension/manual",
		zap.Int64("job_id", job.ID),
		zap.String("article_url", articleURL),
		zap.Time("fulfilled_at", time.Now()),
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
	case publish.JobStatusQueued,
		publish.JobStatusSubmitting,
		publish.JobStatusPolling,
		publish.JobStatusAwaitingExtension,
		publish.JobStatusAwaitingManual:
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

// GetArticleURL returns the article URL for a successful publish job.
func (s *PublishService) GetArticleURL(ctx context.Context, jobID int64) (string, error) {
	records, err := s.recordRepo.FindByPublishJobID(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("publishService.GetArticleURL: %w", err)
	}
	for _, r := range records {
		if r.ArticleURL != nil && *r.ArticleURL != "" {
			return *r.ArticleURL, nil
		}
	}
	return "", nil
}

// RetryJob resets a failed job to queued status for re-processing.
func (s *PublishService) RetryJob(ctx context.Context, publicID string) (*publish.PublishJob, error) {
	job, err := s.jobRepo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("publishService.RetryJob: %w", err)
	}
	if job == nil {
		return nil, ErrNotFound
	}

	if job.Status != publish.JobStatusFailed {
		return nil, fmt.Errorf("publishService.RetryJob: only failed jobs can be retried, current status: %s", job.Status)
	}

	job.Status = publish.JobStatusQueued
	job.LastError = nil
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("publishService.RetryJob: %w", err)
	}

	s.logger.Info("publish job retried",
		zap.String("public_id", publicID),
		zap.Int("retry_count", job.RetryCount),
	)
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

	// Load the real draft so we send actual title/content/digest to WeChat instead
	// of placeholder text. Without this the published article body is "Stub content".
	d, err := s.draftRepo.FindByID(ctx, job.DraftID)
	if err != nil {
		return s.failJob(ctx, job, fmt.Sprintf("load draft: %v", err))
	}
	if d == nil {
		return s.failJob(ctx, job, fmt.Sprintf("draft %d not found", job.DraftID))
	}
	if d.CompiledHTML == "" {
		return s.failJob(ctx, job, "draft compiled_html is empty (run html_compile first)")
	}

	article := adapter.WeChatArticle{
		Title:   d.Title,
		Author:  d.AuthorName,
		Digest:  d.Digest,
		Content: d.CompiledHTML,
	}
	if d.ContentSourceURL != nil {
		article.SourceURL = *d.ContentSourceURL
	}

	// WeChat requires thumb_media_id for every draft; auto-generate if the user
	// hasn't picked one. Failures here are surfaced because a publish without a
	// thumb will be rejected by /draft/add anyway.
	if s.coverImageSvc != nil {
		coverURL, err := s.coverImageSvc.EnsureCover(ctx, d)
		if err != nil {
			return s.failJob(ctx, job, fmt.Sprintf("ensure cover: %v", err))
		}
		article.ThumbURL = coverURL
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
