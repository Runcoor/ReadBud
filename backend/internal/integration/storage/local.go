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
	rootDir    string // resolved absolute path; symlinks resolved when possible
	publicBase string
	logger     *zap.Logger
}

// NewLocalStorageProvider creates a new LocalStorageProvider.
// rootDir is resolved to an absolute, symlink-free path at construction so that
// security checks remain stable across the process lifetime regardless of CWD.
// If the directory does not exist yet, only lexical absolute resolution is used.
func NewLocalStorageProvider(rootDir, publicBase string, logger *zap.Logger) *LocalStorageProvider {
	resolved, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		// Directory may not exist yet (Upload will create it). Fall back to lexical Abs.
		abs, absErr := filepath.Abs(rootDir)
		if absErr != nil {
			logger.Warn("localStorage: failed to resolve rootDir, using as-is",
				zap.String("root_dir", rootDir),
				zap.Error(absErr),
			)
			resolved = rootDir
		} else {
			resolved = abs
		}
	}
	return &LocalStorageProvider{
		rootDir:    resolved,
		publicBase: strings.TrimRight(publicBase, "/"),
		logger:     logger,
	}
}

// absPath returns the on-disk path for (bucket, key) and verifies it stays
// within rootDir even after symlink resolution of any existing parent dirs.
func (s *LocalStorageProvider) absPath(bucket, key string) (string, error) {
	if strings.Contains(bucket, "..") || strings.Contains(key, "..") {
		return "", errors.New("localStorage: bucket/key may not contain '..'")
	}
	full := filepath.Join(s.rootDir, bucket, key)

	// Resolve the deepest existing ancestor of `full` to catch symlink escapes.
	// We walk up the path until we find a component that exists, then EvalSymlinks it.
	check := full
	for {
		resolved, err := filepath.EvalSymlinks(check)
		if err == nil {
			if !pathWithinRoot(resolved, s.rootDir) {
				return "", errors.New("localStorage: resolved path escapes root")
			}
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("localStorage.absPath eval %q: %w", check, err)
		}
		parent := filepath.Dir(check)
		if parent == check {
			// Reached filesystem root without finding any existing ancestor; reject.
			return "", errors.New("localStorage: no existing ancestor for path")
		}
		check = parent
	}

	// Final lexical sanity check against the (already symlink-free) rootDir.
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("localStorage.absPath: %w", err)
	}
	if !pathWithinRoot(fullAbs, s.rootDir) {
		return "", errors.New("localStorage: lexical path escapes root")
	}
	return full, nil
}

func pathWithinRoot(p, root string) bool {
	if p == root {
		return true
	}
	return strings.HasPrefix(p, root+string(os.PathSeparator))
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

// GetURL returns the public URL for an existing object (no presigning, no existence check).
func (s *LocalStorageProvider) GetURL(ctx context.Context, bucket, key string) (string, error) {
	if _, err := s.absPath(bucket, key); err != nil {
		return "", err
	}
	return s.publicURL(bucket, key), nil
}

// Download returns the bytes for the object at <rootDir>/<bucket>/<key>.
// Wraps os.ErrNotExist when the object is missing so callers can
// distinguish missing-file from other failures via errors.Is.
func (s *LocalStorageProvider) Download(ctx context.Context, bucket, key string) ([]byte, error) {
	p, err := s.absPath(bucket, key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("localStorage.Download: %w", err)
	}
	return data, nil
}

// Delete removes the object. A missing file is not treated as an error and produces no log line.
func (s *LocalStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	p, err := s.absPath(bucket, key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
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

var _ adapter.StorageProvider = (*LocalStorageProvider)(nil)
