package snapotter

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
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

func TestPreviewSettingsDependOnOrientation(t *testing.T) {
	t.Parallel()

	portrait := Result{Width: 1313, Height: 3700}
	if portrait.Landscape() {
		t.Fatal("expected 1313x3700 to be portrait")
	}
	if got, want := portrait.PreviewSettings(), PortraitPreviewSettings; got != want {
		t.Fatalf("portrait preview settings mismatch: got %s want %s", got, want)
	}
	if !json.Valid([]byte(portrait.PreviewSettings())) {
		t.Fatal("portrait preview settings is not valid JSON")
	}

	landscape := Result{Width: 3700, Height: 1313}
	if !landscape.Landscape() {
		t.Fatal("expected 3700x1313 to be landscape")
	}
	if got, want := landscape.PreviewSettings(), LandscapePreviewSettings; got != want {
		t.Fatalf("landscape preview settings mismatch: got %s want %s", got, want)
	}
	if !json.Valid([]byte(landscape.PreviewSettings())) {
		t.Fatal("landscape preview settings is not valid JSON")
	}

	square := Result{Width: 800, Height: 800}
	if !square.Landscape() {
		t.Fatal("expected square to be treated as landscape like the dimensions field")
	}
}

func TestPreviewSettingsJSONValues(t *testing.T) {
	t.Parallel()

	for name, settings := range map[string]string{
		"portrait":  PortraitPreviewSettings,
		"landscape": LandscapePreviewSettings,
	} {
		var parsed struct {
			Format   string `json:"format"`
			Quality  int    `json:"quality"`
			MaxWidth int    `json:"maxWidth"`
		}
		if err := json.Unmarshal([]byte(settings), &parsed); err != nil {
			t.Fatalf("%s settings are not valid JSON: %v", name, err)
		}
		if parsed.Format != "webp" {
			t.Fatalf("%s settings format, got %q want webp", name, parsed.Format)
		}
		if want := map[string]int{"portrait": 1200, "landscape": 600}[name]; parsed.MaxWidth != want {
			t.Fatalf("%s maxWidth, got %d want %d", name, parsed.MaxWidth, want)
		}
		if parsed.Quality != 75 {
			t.Fatalf("%s quality, got %d want 75", name, parsed.Quality)
		}
	}
}

func writeTestImage(t *testing.T, width, height int) string {
	t.Helper()

	path := t.TempDir() + "/menu.webp"
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("failed to write test image: %v", err)
	}
	return path
}

func TestSettingsForFilePicksWidthByOrientation(t *testing.T) {
	t.Parallel()

	landscape := writeTestImage(t, 640, 400)
	if got, want := SettingsForFile(landscape, 85, LandscapeFullMaxWidth, PortraitFullMaxWidth), webpSettings(85, 1920); got != want {
		t.Fatalf("landscape file settings mismatch: got %s want %s", got, want)
	}

	portrait := writeTestImage(t, 400, 640)
	if got, want := SettingsForFile(portrait, 85, LandscapeFullMaxWidth, PortraitFullMaxWidth), webpSettings(85, 1200); got != want {
		t.Fatalf("portrait file settings mismatch: got %s want %s", got, want)
	}

	square := writeTestImage(t, 500, 500)
	if got, want := SettingsForFile(square, 85, LandscapeFullMaxWidth, PortraitFullMaxWidth), webpSettings(85, 1920); got != want {
		t.Fatalf("square file settings mismatch: got %s want %s", got, want)
	}
}

func TestSettingsForFileFallsBackToPortraitOnUnreadableFile(t *testing.T) {
	t.Parallel()

	missing := t.TempDir() + "/missing.webp"
	if got, want := SettingsForFile(missing, 85, LandscapeFullMaxWidth, PortraitFullMaxWidth), webpSettings(85, 1200); got != want {
		t.Fatalf("missing file settings mismatch: got %s want %s", got, want)
	}

	unreadable := t.TempDir() + "/garbage.webp"
	if err := os.WriteFile(unreadable, []byte("not an image"), 0o644); err != nil {
		t.Fatalf("failed to write garbage file: %v", err)
	}
	if got, want := SettingsForFile(unreadable, 85, LandscapeFullMaxWidth, PortraitFullMaxWidth), webpSettings(85, 1200); got != want {
		t.Fatalf("garbage file settings mismatch: got %s want %s", got, want)
	}
}
