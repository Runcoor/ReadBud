package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// LocalStorageProvider stores objects on the local filesystem.
// URLs returned point at a public-facing static route (e.g., "/static/images/...")
// served by the API process via gin.Static.
type LocalStorageProvider struct {
	rootDir    string
	publicBase string
	logger     *zap.Logger
}

// NewLocalStorageProvider creates a new LocalStorageProvider.
// rootDir is the absolute or relative directory on disk where files are written.
// publicBase is the URL prefix prepended to (bucket/key) when returning URLs
// (e.g., "/static/images" or "https://cdn.example.com/img").
func NewLocalStorageProvider(rootDir, publicBase string, logger *zap.Logger) *LocalStorageProvider {
	return &LocalStorageProvider{
		rootDir:    rootDir,
		publicBase: strings.TrimRight(publicBase, "/"),
		logger:     logger,
	}
}

func (s *LocalStorageProvider) absPath(bucket, key string) (string, error) {
	if strings.Contains(bucket, "..") || strings.Contains(key, "..") {
		return "", errors.New("localStorage: bucket/key may not contain '..'")
	}
	full := filepath.Join(s.rootDir, bucket, key)
	rootAbs, err := filepath.Abs(s.rootDir)
	if err != nil {
		return "", fmt.Errorf("localStorage.absPath: %w", err)
	}
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("localStorage.absPath: %w", err)
	}
	if !strings.HasPrefix(fullAbs, rootAbs+string(os.PathSeparator)) && fullAbs != rootAbs {
		return "", errors.New("localStorage: resolved path escapes root")
	}
	return full, nil
}

// Upload writes data to <rootDir>/<bucket>/<key> and returns its public URL.
func (s *LocalStorageProvider) Upload(ctx context.Context, bucket, key string, data []byte, contentType string) (string, error) {
	p, err := s.absPath(bucket, key)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", fmt.Errorf("localStorage.Upload: mkdir: %w", err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		return "", fmt.Errorf("localStorage.Upload: write: %w", err)
	}
	url := s.publicURL(bucket, key)
	s.logger.Info("local storage: uploaded object",
		zap.String("bucket", bucket),
		zap.String("key", key),
		zap.Int("size_bytes", len(data)),
		zap.String("content_type", contentType),
		zap.String("url", url),
	)
	return url, nil
}

// GetURL returns the public URL for an existing object (no presigning).
func (s *LocalStorageProvider) GetURL(ctx context.Context, bucket, key string) (string, error) {
	if _, err := s.absPath(bucket, key); err != nil {
		return "", err
	}
	return s.publicURL(bucket, key), nil
}

// Delete removes the object. A missing file is not treated as an error.
func (s *LocalStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	p, err := s.absPath(bucket, key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("localStorage.Delete: %w", err)
	}
	s.logger.Info("local storage: deleted object",
		zap.String("bucket", bucket),
		zap.String("key", key),
	)
	return nil
}

func (s *LocalStorageProvider) publicURL(bucket, key string) string {
	return s.publicBase + "/" + path.Join(bucket, filepath.ToSlash(key))
}

// Compile-time check.
var _ adapter.StorageProvider = (*LocalStorageProvider)(nil)
