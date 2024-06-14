package files

import (
	"path/filepath"
	"strings"
)

func GetPathInformation(path string) (dir string, name string, ext string) {
	dir = filepath.Dir(path)
	base := filepath.Base(path)
	ext = filepath.Ext(base)
	return dir, strings.TrimSuffix(base, ext), ext
}
