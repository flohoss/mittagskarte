package hash

import (
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"os"
)

func HashFile(filePath string) []byte {
	f, err := os.Open(filePath)
	if err != nil {
		slog.Error(err.Error())
		return []byte{}
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		slog.Error(err.Error())
		return []byte{}
	}

	return h.Sum(nil)
}

func AddHashQueryToFileName(filePath string) string {
	return fmt.Sprintf("/%s?hash=%x", filePath, HashFile(filePath))
}
