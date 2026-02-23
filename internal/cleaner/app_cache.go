package cleaner

import (
	"os"
	"path/filepath"
)

type AppCacheCleaner struct{}

func (c *AppCacheCleaner) Name() string {
	return "App Specific Caches"
}

func (c *AppCacheCleaner) RequiresRoot() bool {
	return false
}

func (c *AppCacheCleaner) getCachePaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(home, ".config")

	// Target heavy electron apps that usually bury Cache in ~/.config/<app>
	return []string{
		filepath.Join(configDir, "discord", "Cache"),
		filepath.Join(configDir, "discord", "Code Cache"),
		filepath.Join(configDir, "discord", "DawnCache"),

		filepath.Join(configDir, "Slack", "Cache"),
		filepath.Join(configDir, "Slack", "Code Cache"),
		filepath.Join(configDir, "Slack", "Service Worker", "CacheStorage"),

		filepath.Join(configDir, "Code", "Cache"),
		filepath.Join(configDir, "Code", "CachedData"),
		filepath.Join(configDir, "Code", "CachedExtensionVSIXs"),

		filepath.Join(configDir, "spotify", "Storage"),
	}, nil
}

func (c *AppCacheCleaner) Scan() (int64, error) {
	paths, err := c.getCachePaths()
	if err != nil {
		return 0, err
	}

	var size int64
	for _, p := range paths {
		s, _ := simpleDirScan(p)
		size += s
	}
	return size, nil
}

func (c *AppCacheCleaner) Clean() error {
	paths, err := c.getCachePaths()
	if err != nil {
		return err
	}

	for _, p := range paths {
		_ = os.RemoveAll(p)
	}
	return nil
}
