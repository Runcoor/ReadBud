// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package imageresize

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"net/http"
	"testing"
)

const oneMB = 1 * 1024 * 1024

// makeNoisyPNG produces a w×h PNG full of pseudo-random pixels. Random
// pixels defeat compression so the encoded byte size is roughly 4×w×h
// before deflate, which is exactly what we need to exceed the 1 MB
// budget without depending on real fixtures.
func makeNoisyPNG(t *testing.T, w, h int, seed int64) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r := rand.New(rand.NewSource(seed))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(r.Intn(256)),
				G: uint8(r.Intn(256)),
				B: uint8(r.Intn(256)),
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	return buf.Bytes()
}

func makeSolidJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode test jpeg: %v", err)
	}
	return buf.Bytes()
}

func TestFitWeChat_SmallPNGUnchanged(t *testing.T) {
	// 8×8 noise PNG — well under any sensible budget.
	src := makeNoisyPNG(t, 8, 8, 1)
	if len(src) >= oneMB {
		t.Fatalf("test fixture unexpectedly large: %d bytes", len(src))
	}

	out, mime, err := FitWeChat(src, oneMB)
	if err != nil {
		t.Fatalf("FitWeChat: %v", err)
	}
	if !bytes.Equal(out, src) {
		t.Fatalf("expected pass-through, got different bytes (in=%d out=%d)", len(src), len(out))
	}
	if mime != "image/png" {
		t.Fatalf("expected image/png, got %q", mime)
	}
}

func TestFitWeChat_SmallJPEGUnchanged(t *testing.T) {
	src := makeSolidJPEG(t, 64, 64)
	out, mime, err := FitWeChat(src, oneMB)
	if err != nil {
		t.Fatalf("FitWeChat: %v", err)
	}
	if !bytes.Equal(out, src) {
		t.Fatalf("expected pass-through")
	}
	if mime != "image/jpeg" {
		t.Fatalf("expected image/jpeg, got %q", mime)
	}
}

func TestFitWeChat_OversizedPNGRecodedAsJPEG(t *testing.T) {
	// 1024×768 noise PNG → ~3 MB encoded; well over 1 MB.
	src := makeNoisyPNG(t, 1024, 768, 42)
	if len(src) <= oneMB {
		t.Fatalf("test fixture not oversized: %d bytes", len(src))
	}

	out, mime, err := FitWeChat(src, oneMB)
	if err != nil {
		t.Fatalf("FitWeChat: %v", err)
	}
	if len(out) > oneMB {
		t.Fatalf("output exceeds budget: %d > %d", len(out), oneMB)
	}
	if mime != "image/jpeg" {
		t.Fatalf("expected JPEG, got %q", mime)
	}
	if got := http.DetectContentType(out); got != "image/jpeg" {
		t.Fatalf("DetectContentType disagrees: got %q", got)
	}
	// Sanity: the compressed result must still decode.
	if _, _, err := image.Decode(bytes.NewReader(out)); err != nil {
		t.Fatalf("output failed to decode: %v", err)
	}
}

func TestFitWeChat_HugeImageScaledDown(t *testing.T) {
	// 2048×1536 noise PNG → ~12 MB encoded. Even at JPEG q=55 a noise
	// image at full resolution will overshoot 1 MB, so this exercises
	// the scale-down path.
	src := makeNoisyPNG(t, 2048, 1536, 7)
	if len(src) <= oneMB {
		t.Fatalf("test fixture not oversized: %d bytes", len(src))
	}

	out, mime, err := FitWeChat(src, oneMB)
	if err != nil {
		t.Fatalf("FitWeChat: %v", err)
	}
	if len(out) > oneMB {
		t.Fatalf("output exceeds budget: %d > %d", len(out), oneMB)
	}
	if mime != "image/jpeg" {
		t.Fatalf("expected JPEG, got %q", mime)
	}
	decoded, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("output failed to decode: %v", err)
	}
	// We expect the dimensions to have shrunk relative to the source.
	b := decoded.Bounds()
	if b.Dx() >= 2048 || b.Dy() >= 1536 {
		t.Fatalf("expected scaled-down output, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestFitWeChat_EmptyInput(t *testing.T) {
	if _, _, err := FitWeChat(nil, oneMB); err == nil {
		t.Fatalf("expected error for empty input")
	}
}

func TestFitWeChat_InvalidBytes(t *testing.T) {
	if _, _, err := FitWeChat([]byte("not an image"), oneMB); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestFitWeChat_InvalidBudget(t *testing.T) {
	src := makeSolidJPEG(t, 8, 8)
	if _, _, err := FitWeChat(src, 0); err == nil {
		t.Fatalf("expected error for maxBytes=0")
	}
}
