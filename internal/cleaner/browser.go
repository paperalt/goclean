package cleaner

import (
	"os"
	"path/filepath"
)

type BrowserCleaner struct{}

func (c *BrowserCleaner) Name() string {
	return "Browser Caches"
}

func (c *BrowserCleaner) RequiresRoot() bool {
	return false
}

func (c *BrowserCleaner) Scan() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	targets := []string{
		filepath.Join(home, ".cache", "google-chrome"),
		filepath.Join(home, ".cache", "chromium"),
		filepath.Join(home, ".cache", "mozilla", "firefox"),
		filepath.Join(home, ".cache", "BraveSoftware"),
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

func (c *BrowserCleaner) Clean() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// For browsers, we usually want to delete the Cache folder inside the profile
	// But simply verifying the detailed structure is hard (random profile names).
	// Deleting ~/.cache/google-chrome is usually safe (it regenerates).
	
	targets := []string{
		filepath.Join(home, ".cache", "google-chrome"),
		filepath.Join(home, ".cache", "chromium"),
		// Firefox is tricky: ~/.cache/mozilla/firefox/PROFILE/cache2
		// We can try to walk and find "cache2" dirs?
		// Or just nuke ~/.cache/mozilla/firefox which contains cache data usually separated from user profile data (which is in ~/.mozilla)
		filepath.Join(home, ".cache", "mozilla", "firefox"), 
		filepath.Join(home, ".cache", "BraveSoftware"),
	}

	for _, t := range targets {
		os.RemoveAll(t)
	}
	return nil
}
