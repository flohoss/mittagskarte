package checksum

import (
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"os"
)

func ChecksumFile(filePath string) uint32 {
	f, err := os.Open(filePath)
	if err != nil {
		slog.Error(err.Error())
		return 0
	}
	defer f.Close()

	h := crc32.NewIEEE()
	if _, err := io.Copy(h, f); err != nil {
		slog.Error(err.Error())
		return 0
	}

	return h.Sum32()
}

func SuffixQuery(filePath string) string {
	return fmt.Sprintf("/%s?checksum=%x", filePath, ChecksumFile(filePath))
}
