package fsutil

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/tools/filesystem"
)

type customReader struct{}

type customReadSeekCloser struct {
	*bytes.Reader
}

func (c customReadSeekCloser) Close() error {
	return nil
}

func (customReader) Open() (io.ReadSeekCloser, error) {
	return customReadSeekCloser{Reader: bytes.NewReader(nil)}, nil
}

func TestLocalPathWithPathReader(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	source := filepath.Join(dir, "menu.webp")
	if err := os.WriteFile(source, []byte("menu"), 0o644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	f, err := filesystem.NewFileFromPath(source)
	if err != nil {
		t.Fatalf("failed to build filesystem file: %v", err)
	}

	gotPath, cleanup, err := LocalPath(f, dir)
	if err != nil {
		t.Fatalf("LocalPath returned error: %v", err)
	}

	if gotPath != source {
		t.Fatalf("unexpected path, got %q want %q", gotPath, source)
	}

	cleanup()

	if _, err := os.Stat(source); err != nil {
		t.Fatalf("expected source file to remain after cleanup, got: %v", err)
	}
}

func TestLocalPathWithMultipartReader(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := []byte("multipart menu payload")

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "menu.txt")
	if err != nil {
		t.Fatalf("failed to create multipart part: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}

	_, header, err := req.FormFile("file")
	if err != nil {
		t.Fatalf("failed to get multipart file header: %v", err)
	}

	f, err := filesystem.NewFileFromMultipart(header)
	if err != nil {
		t.Fatalf("failed to build filesystem file: %v", err)
	}

	gotPath, cleanup, err := LocalPath(f, dir)
	if err != nil {
		t.Fatalf("LocalPath returned error: %v", err)
	}

	data, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("failed to read copied multipart file: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("unexpected copied content, got %q want %q", data, content)
	}

	cleanup()

	if _, err := os.Stat(gotPath); !os.IsNotExist(err) {
		t.Fatalf("expected cleanup to remove temp file %q", gotPath)
	}
}

func TestLocalPathUnsupportedReader(t *testing.T) {
	t.Parallel()

	_, cleanup, err := LocalPath(&filesystem.File{Reader: customReader{}}, t.TempDir())
	if err == nil {
		t.Fatal("expected unsupported reader error, got nil")
	}
	if cleanup != nil {
		t.Fatal("expected nil cleanup for unsupported reader")
	}
	if !strings.Contains(err.Error(), "unsupported file reader type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
