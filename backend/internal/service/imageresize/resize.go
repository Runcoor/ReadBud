// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Package imageresize compresses and downscales images so they fit
// hard size limits set by downstream APIs (specifically the WeChat
// content image endpoint, which caps each image at 1 MB and only
// accepts jpg/png).
//
// The transform is intentionally lossy: when an image cannot fit at
// its source resolution, the function progressively re-encodes as JPEG
// at decreasing quality, then halves the dimensions, until it fits or
// runs out of attempts.
package imageresize

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"

	// Register the PNG decoder so image.Decode can sniff PNG inputs.
	// JPEG decoder is registered transitively by the named jpeg import above.
	_ "image/png"

	"golang.org/x/image/draw"
)

const (
	mimePNG  = "image/png"
	mimeJPEG = "image/jpeg"
)

// FitWeChat returns image bytes that are ≤ maxBytes, valid PNG or JPEG,
// and the matching MIME type. If the input already satisfies both
// conditions it is returned unchanged.
//
// When the input is too large, the function first tries JPEG re-encoding
// at decreasing quality at the source resolution. If that still exceeds
// the budget, it scales the image down (Catmull-Rom resampling) and
// re-tries. Returns an error if every attempt overshoots the budget.
func FitWeChat(data []byte, maxBytes int) ([]byte, string, error) {
	if len(data) == 0 {
		return nil, "", errors.New("imageresize: empty data")
	}
	if maxBytes <= 0 {
		return nil, "", fmt.Errorf("imageresize: maxBytes must be > 0, got %d", maxBytes)
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("imageresize: decode: %w", err)
	}

	// Fast path: already under budget and a supported format.
	if len(data) <= maxBytes {
		switch format {
		case "png":
			return data, mimePNG, nil
		case "jpeg":
			return data, mimeJPEG, nil
		}
		// Other formats (gif, webp, ...) — fall through and re-encode as JPEG.
	}

	// Re-encode strategy. Try larger scales / higher quality first so we
	// keep as much fidelity as possible.
	scales := []float64{1.0, 0.85, 0.70, 0.55, 0.40}
	qualities := []int{85, 75, 65, 55}

	var bestErr error
	for _, scale := range scales {
		scaled := scaleImage(img, scale)
		for _, q := range qualities {
			var buf bytes.Buffer
			if encErr := jpeg.Encode(&buf, scaled, &jpeg.Options{Quality: q}); encErr != nil {
				bestErr = fmt.Errorf("encode q=%d scale=%.2f: %w", q, scale, encErr)
				continue
			}
			if buf.Len() <= maxBytes {
				return buf.Bytes(), mimeJPEG, nil
			}
		}
	}

	if bestErr != nil {
		return nil, "", fmt.Errorf("imageresize: cannot fit under %d bytes: %w", maxBytes, bestErr)
	}
	return nil, "", fmt.Errorf("imageresize: cannot fit under %d bytes after %d attempts",
		maxBytes, len(scales)*len(qualities))
}

// scaleImage returns src downscaled by the given factor. For scale ≥ 1.0
// it returns src unchanged. The output is an *image.RGBA produced by
// Catmull-Rom resampling, which gives noticeably better quality than
// nearest-neighbour at the cost of some CPU.
func scaleImage(src image.Image, scale float64) image.Image {
	if scale >= 1.0 {
		return src
	}
	sb := src.Bounds()
	dw := int(float64(sb.Dx()) * scale)
	dh := int(float64(sb.Dy()) * scale)
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, sb, draw.Over, nil)
	return dst
}

