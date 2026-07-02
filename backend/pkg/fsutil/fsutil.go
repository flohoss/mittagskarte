package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func LocalPath(f *filesystem.File, dir string) (path string, cleanup func(), err error) {
	switch reader := f.Reader.(type) {
	case *filesystem.PathReader:
		return reader.Path, func() {}, nil
	case *filesystem.MultipartReader:
		rc, err := reader.Open()
		if err != nil {
			return "", nil, err
		}
		ext := filepath.Ext(f.Name)
		tmp := filepath.Join(dir, fmt.Sprintf("upload_%d%s", time.Now().UnixNano(), ext))
		out, err := os.Create(tmp)
		if err != nil {
			rc.Close()
			return "", nil, err
		}
		if _, err = io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			os.Remove(tmp)
			return "", nil, err
		}
		out.Close()
		rc.Close()
		return tmp, func() { os.Remove(tmp) }, nil
	default:
		return "", nil, fmt.Errorf("unsupported file reader type %T", f.Reader)
	}
}
