package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
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
