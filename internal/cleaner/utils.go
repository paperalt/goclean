package cleaner

import (
	"os"
	"path/filepath"
)

// simpleDirScan walks a directory and returns its total size
func simpleDirScan(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
