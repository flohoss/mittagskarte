package files

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
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
