package service

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// AuditAction defines the type of auditable operation.
type AuditAction string

// Audit action constants for trackable operations.
const (
	AuditActionTaskCreate      AuditAction = "task.create"
	AuditActionTaskRetry       AuditAction = "task.retry"
	AuditActionDraftUpdate     AuditAction = "draft.update"
	AuditActionBlockUpdate     AuditAction = "draft.block.update"
	AuditActionPublishCreate   AuditAction = "publish.create"
	AuditActionPublishRetry    AuditAction = "publish.retry"
	AuditActionPublishCancel   AuditAction = "publish.cancel"
	AuditActionProviderCreate  AuditAction = "provider.create"
	AuditActionProviderUpdate  AuditAction = "provider.update"
	AuditActionWechatCreate    AuditAction = "wechat.create"
	AuditActionWechatUpdate    AuditAction = "wechat.update"
	AuditActionTopicCreate     AuditAction = "topic.create"
	AuditActionTopicUpdate     AuditAction = "topic.update"
	AuditActionTopicDelete     AuditAction = "topic.delete"
	AuditActionDistGenerate    AuditAction = "distribution.generate"
	AuditActionDistDelete      AuditAction = "distribution.delete"
	AuditActionMetricsSync     AuditAction = "metrics.sync"
	AuditActionLogin           AuditAction = "auth.login"
	AuditActionTokenRefresh    AuditAction = "auth.refresh"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	Timestamp  time.Time         `json:"timestamp"`
	RequestID  string            `json:"request_id"`
	Action     AuditAction       `json:"action"`
	UserID     int64             `json:"user_id,omitempty"`
	Username   string            `json:"username,omitempty"`
	ResourceID string            `json:"resource_id,omitempty"`
	Detail     string            `json:"detail,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Success    bool              `json:"success"`
	Error      string            `json:"error,omitempty"`
}

// AuditService provides structured audit logging for business operations.
// In a production system, audit entries would be persisted to a database table.
// This implementation uses structured Zap logging for audit trail.
type AuditService struct {
	logger *zap.Logger
}

// NewAuditService creates a new AuditService.
func NewAuditService(logger *zap.Logger) *AuditService {
	return &AuditService{
		logger: logger.Named("audit"),
	}
}

// Record logs an audit entry with structured fields.
func (s *AuditService) Record(_ context.Context, entry AuditEntry) {
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	fields := []zap.Field{
		zap.Time("audit_timestamp", entry.Timestamp),
		zap.String("request_id", entry.RequestID),
		zap.String("action", string(entry.Action)),
		zap.Bool("success", entry.Success),
	}

	if entry.UserID > 0 {
		fields = append(fields, zap.Int64("user_id", entry.UserID))
	}
	if entry.Username != "" {
		fields = append(fields, zap.String("username", entry.Username))
	}
	if entry.ResourceID != "" {
		fields = append(fields, zap.String("resource_id", entry.ResourceID))
	}
	if entry.Detail != "" {
		fields = append(fields, zap.String("detail", entry.Detail))
	}
	if entry.Error != "" {
		fields = append(fields, zap.String("error", entry.Error))
	}
	if len(entry.Metadata) > 0 {
		for k, v := range entry.Metadata {
			fields = append(fields, zap.String("meta_"+k, v))
		}
	}

	if entry.Success {
		s.logger.Info("audit_event", fields...)
	} else {
		s.logger.Warn("audit_event", fields...)
	}
}

// RecordSuccess is a convenience method for successful operations.
func (s *AuditService) RecordSuccess(ctx context.Context, action AuditAction, requestID string, userID int64, username string, resourceID string, detail string) {
	s.Record(ctx, AuditEntry{
		Action:     action,
		RequestID:  requestID,
		UserID:     userID,
		Username:   username,
		ResourceID: resourceID,
		Detail:     detail,
		Success:    true,
	})
}

// RecordFailure is a convenience method for failed operations.
func (s *AuditService) RecordFailure(ctx context.Context, action AuditAction, requestID string, userID int64, username string, resourceID string, errMsg string) {
	s.Record(ctx, AuditEntry{
		Action:     action,
		RequestID:  requestID,
		UserID:     userID,
		Username:   username,
		ResourceID: resourceID,
		Success:    false,
		Error:      errMsg,
	})
}
