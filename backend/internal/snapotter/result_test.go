package snapotter

import (
	"os"
	"testing"
)

func TestNewResultDecodesDimensions(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/pixel.webp")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	result, err := newResult(data, "menu.webp")
	if err != nil {
		t.Fatalf("newResult returned error: %v", err)
	}
	if result.Width != 8 || result.Height != 6 {
		t.Fatalf("unexpected dimensions, got %dx%d want 8x6", result.Width, result.Height)
	}
	if result.Name != "menu.webp" {
		t.Fatalf("unexpected name, got %q", result.Name)
	}
	if string(result.Data) != string(data) {
		t.Fatal("expected result data to carry the input bytes")
	}
}

func TestNewResultRejectsInvalidImage(t *testing.T) {
	t.Parallel()

	_, err := newResult([]byte("definitely not an image"), "menu.webp")
	if err == nil {
		t.Fatal("expected error for invalid image data, got nil")
	}
}
