package restaurant

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestLastCheckFromError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus LastCheckStatus
		wantDetail string
	}{
		{name: "nil", err: nil, wantStatus: LastCheckStatusSuccess, wantDetail: ""},
		{name: "manual upload only", err: ErrManualUploadOnly, wantStatus: LastCheckStatusNotChanged, wantDetail: ""},
		{name: "menu unchanged", err: ErrMenuUnchanged, wantStatus: LastCheckStatusNotChanged, wantDetail: ""},
		{name: "generic error", err: errors.New("boom"), wantStatus: LastCheckStatusError, wantDetail: "boom"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, detail := LastCheckFromError(tc.err)
			if status != tc.wantStatus {
				t.Fatalf("unexpected status, got %q want %q", status, tc.wantStatus)
			}
			if detail != tc.wantDetail {
				t.Fatalf("unexpected detail, got %q want %q", detail, tc.wantDetail)
			}
		})
	}
}

func TestReadMenuDimensions(t *testing.T) {
	t.Parallel()

	t.Run("landscape image", func(t *testing.T) {
		t.Parallel()

		filePath := writePNG(t, 4, 2)
		dims, err := readMenuDimensions(filePath)
		if err != nil {
			t.Fatalf("readMenuDimensions returned error: %v", err)
		}

		if dims.Width != 4 || dims.Height != 2 || !dims.Landscape {
			t.Fatalf("unexpected dimensions: %#v", dims)
		}
	})

	t.Run("portrait image", func(t *testing.T) {
		t.Parallel()

		filePath := writePNG(t, 2, 4)
		dims, err := readMenuDimensions(filePath)
		if err != nil {
			t.Fatalf("readMenuDimensions returned error: %v", err)
		}

		if dims.Width != 2 || dims.Height != 4 || dims.Landscape {
			t.Fatalf("unexpected dimensions: %#v", dims)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		if _, err := readMenuDimensions(filepath.Join(t.TempDir(), "missing.png")); err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})
}

func writePNG(t *testing.T, width int, height int) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	filePath := filepath.Join(t.TempDir(), "menu.png")
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create temp image: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	return filePath
}
