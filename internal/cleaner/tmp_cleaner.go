package cleaner

import (
	"os"
	"path/filepath"
)

type TmpCleaner struct{}

func (c *TmpCleaner) Name() string {
	return "System Temp Files"
}

func (c *TmpCleaner) RequiresRoot() bool {
	return true
}

func (c *TmpCleaner) Scan() (int64, error) {
	paths := []string{"/tmp", "/var/tmp", "/var/crash"}
	var size int64

	for _, p := range paths {
		err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip permission denied errors
			}
			if !info.IsDir() {
				// Avoid socket and device files
				if info.Mode().IsRegular() {
					size += info.Size()
				}
			}
			return nil
		})
		if err != nil {
			continue
		}
	}
	return size, nil
}

func (c *TmpCleaner) Clean() error {
	paths := []string{"/tmp", "/var/tmp", "/var/crash"}

	for _, p := range paths {
		entries, err := os.ReadDir(p)
		if err != nil {
			continue // Skip if we can't read dir
		}
		for _, entry := range entries {
			// Do not remove important system lock files or X11 socket dirs if possible,
			// though typical `/tmp` wipes just blind delete. `os.RemoveAll` will fail
			// on items actively locked by the OS, which is generally safe enough.

			// Be slightly cautious, skip hidden files starting with .X directly in /tmp,
			// e.g. .X11-unix
			if p == "/tmp" && len(entry.Name()) > 1 && entry.Name()[0:2] == ".X" {
				continue
			}

			fullPath := filepath.Join(p, entry.Name())
			_ = os.RemoveAll(fullPath) // Ignore errors (like permission denied or in-use)
		}
	}
	return nil
}
