package service

import (
	"testing"

	"readbud/internal/adapter"
)

func TestValidateContentImage_Empty(t *testing.T) {
	err := ValidateContentImage(nil, "test.png")
	if err == nil {
		t.Fatal("expected error for empty image data")
	}
}

func TestValidateContentImage_TooLarge(t *testing.T) {
	// Create data exceeding 1MB
	data := make([]byte, adapter.ContentImageMaxBytes+1)
	err := ValidateContentImage(data, "big.png")
	if err == nil {
		t.Fatal("expected error for oversized image")
	}
}

func TestValidateContentImage_InvalidType(t *testing.T) {
	// GIF header
	data := []byte("GIF89a")
	err := ValidateContentImage(data, "test.gif")
	if err == nil {
		t.Fatal("expected error for non-jpg/png image")
	}
}

func TestValidateContentImage_ValidPNG(t *testing.T) {
	// Minimal PNG header
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	err := ValidateContentImage(data, "test.png")
	if err != nil {
		t.Fatalf("unexpected error for valid PNG: %v", err)
	}
}

func TestValidateContentImage_ValidJPEG(t *testing.T) {
	// Minimal JPEG header
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	err := ValidateContentImage(data, "test.jpg")
	if err != nil {
		t.Fatalf("unexpected error for valid JPEG: %v", err)
	}
}

func TestExtractFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"images/2026/03/photo.png", "photo.png"},
		{"photo.png", "photo.png"},
		{"a/b/c/d.jpg", "d.jpg"},
		{"", ""},
	}

	for _, tc := range tests {
		result := extractFilename(tc.input)
		if result != tc.expected {
			t.Errorf("extractFilename(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestIsContentImageType(t *testing.T) {
	tests := []struct {
		assetType string
		expected  bool
	}{
		{"content_image", true},
		{"chart", true},
		{"generated_image", true},
		{"cover_image", false},
		{"unknown", false},
	}

	for _, tc := range tests {
		result := isContentImageType(tc.assetType)
		if result != tc.expected {
			t.Errorf("isContentImageType(%q) = %v, want %v", tc.assetType, result, tc.expected)
		}
	}
}
