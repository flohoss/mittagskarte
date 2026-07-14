package checksum

import (
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestReader(t *testing.T) {
	t.Parallel()

	content := "hello mittagskarte"
	got, err := Reader(strings.NewReader(content))
	if err != nil {
		t.Fatalf("Reader returned error: %v", err)
	}

	want := fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(content)))
	if got != want {
		t.Fatalf("unexpected checksum, got %q want %q", got, want)
	}
}

func TestFile(t *testing.T) {
	t.Parallel()

	t.Run("existing file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "menu.txt")
		content := []byte("daily menu")
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		got, err := File(path)
		if err != nil {
			t.Fatalf("File returned error: %v", err)
		}

		want := fmt.Sprintf("%x", crc32.ChecksumIEEE(content))
		if got != want {
			t.Fatalf("unexpected checksum, got %q want %q", got, want)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		if _, err := File(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("unreadable file", func(t *testing.T) {
		t.Parallel()

		if os.Getuid() == 0 {
			t.Skip("root bypasses file permissions")
		}

		dir := t.TempDir()
		path := filepath.Join(dir, "unreadable.txt")
		if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		if err := os.Chmod(path, 0o000); err != nil {
			t.Fatalf("failed to chmod file: %v", err)
		}
		defer os.Chmod(path, 0o644)

		_, err := File(path)
		if err == nil {
			t.Fatal("expected error from unreadable file, got nil")
		}
	})
}

func TestReaderError(t *testing.T) {
	t.Parallel()

	_, err := Reader(errReader{})
	if err == nil {
		t.Fatal("expected error from failing reader, got nil")
	}
	if err.Error() != "read failed" {
		t.Fatalf("unexpected error, got %v", err)
	}
}
