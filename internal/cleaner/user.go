package cleaner

import (
	"os"
	"path/filepath"
)

type UserCacheCleaner struct{}

func (c *UserCacheCleaner) Name() string {
	return "User Cache (Thumbnails)"
}

func (c *UserCacheCleaner) RequiresRoot() bool {
	return false
}

func (c *UserCacheCleaner) Scan() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	
	// Target just thumbnails for safety MVP
	targets := []string{
		filepath.Join(home, ".cache", "thumbnails"),
	}
	
	var size int64
	for _, t := range targets {
		filepath.Walk(t, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				size += info.Size()
			}
			return nil
		})
	}
	
	return size, nil
}

func (c *UserCacheCleaner) Clean() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	targets := []string{
		filepath.Join(home, ".cache", "thumbnails"),
	}
	
	for _, t := range targets {
		if err := os.RemoveAll(t); err != nil {
			return err
		}
	}
	return nil
}
