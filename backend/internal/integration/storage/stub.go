// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

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

// Download returns a minimal placeholder PNG header so callers can exercise
// the byte-handling code path without a real backing store.
func (s *StubStorageProvider) Download(ctx context.Context, bucket, key string) ([]byte, error) {
	s.logger.Info("stub: downloaded object",
		zap.String("bucket", bucket),
		zap.String("key", key),
	)
	return []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, nil
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
