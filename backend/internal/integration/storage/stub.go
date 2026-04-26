package storage

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// StubStorageProvider is a placeholder implementation of adapter.StorageProvider
// for development and testing. Returns fake URLs.
type StubStorageProvider struct {
	logger *zap.Logger
}

// NewStubStorageProvider creates a new stub storage provider.
func NewStubStorageProvider(logger *zap.Logger) *StubStorageProvider {
	return &StubStorageProvider{logger: logger}
}

// Upload simulates storing data and returns a stub URL.
func (s *StubStorageProvider) Upload(ctx context.Context, bucket, key string, data []byte, contentType string) (string, error) {
	url := fmt.Sprintf("/static/images/%s/%s", bucket, key)
	s.logger.Info("stub: uploaded object",
		zap.String("bucket", bucket),
		zap.String("key", key),
		zap.Int("size_bytes", len(data)),
		zap.String("content_type", contentType),
	)
	return url, nil
}

// GetURL returns a stub presigned URL for the given object.
func (s *StubStorageProvider) GetURL(ctx context.Context, bucket, key string) (string, error) {
	url := fmt.Sprintf("/static/images/%s/%s", bucket, key)
	return url, nil
}

// Delete simulates deleting an object.
func (s *StubStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	s.logger.Info("stub: deleted object",
		zap.String("bucket", bucket),
		zap.String("key", key),
	)
	return nil
}

// Compile-time check that StubStorageProvider satisfies the interface.
var _ adapter.StorageProvider = (*StubStorageProvider)(nil)
