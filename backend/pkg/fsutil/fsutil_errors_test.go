package fsutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func TestLocalPathCreateError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := filepath.Join(root, "readonly")
	if err := os.Mkdir(dir, 0o500); err != nil {
		t.Fatalf("failed to create readonly dir: %v", err)
	}

	f := &filesystem.File{
		Reader: &filesystem.BytesReader{Bytes: []byte("menu")},
		Name:   "menu.webp",
		Size:   4,
	}

	path, cleanup, err := LocalPath(f, dir)
	if err == nil {
		if cleanup != nil {
			cleanup()
		}
		t.Fatal("expected create error in readonly dir, got nil")
	}
	if path != "" {
		t.Fatalf("unexpected path %q", path)
	}
	if cleanup != nil {
		t.Fatal("unexpected cleanup function on error")
	}
}
