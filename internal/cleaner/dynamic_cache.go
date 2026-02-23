package cleaner

import (
	"os"
	"path/filepath"
	"strings"
)

type DynamicCacheCleaner struct{}

func (c *DynamicCacheCleaner) Name() string {
	return "Other User Caches"
}

func (c *DynamicCacheCleaner) RequiresRoot() bool {
	return false
}

func (c *DynamicCacheCleaner) Scan() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	
	cacheDir := filepath.Join(home, ".cache")
	var size int64
	
	// Exclude what we already handle
	excludes := map[string]bool{
		"thumbnails":    true,
		"google-chrome": true,
		"chromium":      true,
		"mozilla":       true,
		"BraveSoftware": true,
	}

	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		// Skip root cache dir
		if path == cacheDir {
			return nil
		}
		
		// Should we skip top level dirs? 
		// If path is ~/.cache/foo, and foo is excluded, skip it.
		rel, _ := filepath.Rel(cacheDir, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		if len(parts) > 0 {
			if excludes[parts[0]] {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, nil
}

func (c *DynamicCacheCleaner) Clean() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cacheDir := filepath.Join(home, ".cache")
	
	excludes := map[string]bool{
		"thumbnails":    true,
		"google-chrome": true,
		"chromium":      true,
		"mozilla":       true,
		"BraveSoftware": true,
	}

	// Walk top level directories in ~/.cache and remove them if not excluded
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !excludes[entry.Name()] {
			path := filepath.Join(cacheDir, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				// Log error?
			}
		}
	}
	return nil
}
