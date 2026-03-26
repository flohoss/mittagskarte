package file

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func Read(folder, filename string, out any) error {
	if out == nil {
		return fmt.Errorf("out must be a non-nil pointer")
	}

	path := filepath.Join(folder, filename+".json")
	slog.Debug("Reading file", "path", path)

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	if err := json.Unmarshal(content, out); err != nil {
		return fmt.Errorf("unmarshal %q: %w", path, err)
	}

	slog.Debug("File loaded successfully", "path", path)
	return nil
}
