package checksum

import (
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"os"
)

func ChecksumFile(filePath string) (uint32, error) {
	f, err := os.Open(filePath)
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	defer f.Close()

	h := crc32.NewIEEE()
	if _, err := io.Copy(h, f); err != nil {
		slog.Error(err.Error())
		return 0, err
	}

	return h.Sum32(), nil
}

func Identical(existingChecksum string, newChecksum uint32) bool {
	return existingChecksum == fmt.Sprintf("%x", newChecksum)
}

func SuffixQuery(filePath string) (string, error) {
	checksum, err := ChecksumFile(filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/%s?v=%x", filePath, checksum), nil
}
