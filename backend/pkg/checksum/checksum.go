package checksum

import (
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"os"
)

func Reader(r io.Reader) (string, error) {
	h := crc32.NewIEEE()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum32()), nil
}

func File(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}
	defer f.Close()

	h := crc32.NewIEEE()
	if _, err := io.Copy(h, f); err != nil {
		slog.Error(err.Error())
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum32()), nil
}
