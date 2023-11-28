package helper

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
)

func GenerateHash(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("could not create hash for file", "file", filename, "err", err)
		return "", err
	}

	hasher := sha256.New()
	hasher.Write(data)
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash, nil
}

func RemoveAllOtherFiles(filePath string) error {
	dir := filepath.Dir(filePath)

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() != filepath.Base(filePath) {

			fileToRemove := filepath.Join(dir, file.Name())
			if err := os.Remove(fileToRemove); err != nil {
				return err
			}
			slog.Debug("Removed file", "file", fileToRemove)
		}
	}

	return nil
}
