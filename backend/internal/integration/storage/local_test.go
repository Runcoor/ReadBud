// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLocalStorage_UploadAndGetURL(t *testing.T) {
	tmp := t.TempDir()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())

	ctx := context.Background()
	url, err := p.Upload(ctx, "generated", "202604/abc.png", []byte("hello"), "image/png")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if url != "/static/images/generated/202604/abc.png" {
		t.Fatalf("unexpected url: %q", url)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "generated", "202604", "abc.png"))
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected file contents: %q", string(data))
	}

	got, err := p.GetURL(ctx, "generated", "202604/abc.png")
	if err != nil {
		t.Fatalf("GetURL returned error: %v", err)
	}
	if got != url {
		t.Fatalf("GetURL mismatch: %q vs %q", got, url)
	}
}

func TestLocalStorage_AutoCreatesParentDirs(t *testing.T) {
	tmp := t.TempDir()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())

	_, err := p.Upload(context.Background(), "a", "b/c/d/file.bin", []byte("x"), "application/octet-stream")
	if err != nil {
		t.Fatalf("Upload failed when parent dir missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "a", "b", "c", "d", "file.bin")); err != nil {
		t.Fatalf("file not created at nested path: %v", err)
	}
}

func TestLocalStorage_DeleteMissingIsNoError(t *testing.T) {
	tmp := t.TempDir()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())

	if err := p.Delete(context.Background(), "generated", "does/not/exist.png"); err != nil {
		t.Fatalf("Delete on missing file should not error, got: %v", err)
	}
}

func TestLocalStorage_RejectsPathTraversal(t *testing.T) {
	tmp := t.TempDir()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())

	_, err := p.Upload(context.Background(), "generated", "../escape.png", []byte("x"), "image/png")
	if err == nil {
		t.Fatalf("expected path traversal to be rejected")
	}
}

func TestLocalStorage_RejectsSymlinkEscape(t *testing.T) {
	tmp := t.TempDir()
	outside := t.TempDir()
	// Create a symlink inside tmp that points outside.
	bucketDir := filepath.Join(tmp, "generated")
	if err := os.MkdirAll(bucketDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	linkPath := filepath.Join(bucketDir, "link")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}

	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())
	_, err := p.Upload(context.Background(), "generated", "link/secret.txt", []byte("x"), "text/plain")
	if err == nil {
		t.Fatalf("expected symlink escape to be rejected, but Upload succeeded")
	}
	// Confirm nothing was written outside the root.
	if _, err := os.Stat(filepath.Join(outside, "secret.txt")); err == nil {
		t.Fatalf("file was written outside rootDir: escape succeeded")
	}
}

func TestLocalStorage_DeleteMissingDoesNotLog(t *testing.T) {
	tmp := t.TempDir()
	// Use observer logger to assert no log line was produced.
	core, recorded := zapObserver()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.New(core))

	if err := p.Delete(context.Background(), "generated", "missing.png"); err != nil {
		t.Fatalf("Delete missing should not error: %v", err)
	}
	if recorded.Len() != 0 {
		t.Fatalf("expected no log entries for missing-file delete, got %d: %+v",
			recorded.Len(), recorded.All())
	}
}

// helper for log assertions
func zapObserver() (zapcore.Core, *observer.ObservedLogs) {
	return observer.New(zapcore.InfoLevel)
}

func TestLocalStorage_DownloadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())

	want := []byte("hello world")
	if _, err := p.Upload(context.Background(), "generated", "202604/abc.png", want, "image/png"); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	got, err := p.Download(context.Background(), "generated", "202604/abc.png")
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("Download mismatch: got %q want %q", got, want)
	}
}

func TestLocalStorage_DownloadMissingReturnsNotExist(t *testing.T) {
	tmp := t.TempDir()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())

	_, err := p.Download(context.Background(), "generated", "missing.png")
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected error to wrap os.ErrNotExist, got %v", err)
	}
}

func TestLocalStorage_DownloadRejectsTraversal(t *testing.T) {
	tmp := t.TempDir()
	p := NewLocalStorageProvider(tmp, "/static/images", zap.NewNop())

	_, err := p.Download(context.Background(), "generated", "../escape.png")
	if err == nil {
		t.Fatalf("expected path traversal to be rejected")
	}
}
