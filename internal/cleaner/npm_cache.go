package cleaner

import (
	"os"
	"os/exec"
	"path/filepath"
)

type NpmCacheCleaner struct{}

func (c *NpmCacheCleaner) Name() string {
	return "NPM Cache"
}

func (c *NpmCacheCleaner) RequiresRoot() bool {
	return false
}

func (c *NpmCacheCleaner) Scan() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	npmDir := filepath.Join(home, ".npm")

	// Check if the directory exists
	info, err := os.Stat(npmDir)
	if err != nil || !info.IsDir() {
		return 0, nil
	}

	return simpleDirScan(npmDir)
}

func (c *NpmCacheCleaner) Clean() error {
	// Let npm itself handle the complex clearing logic if it exists
	if _, err := exec.LookPath("npm"); err == nil {
		cmd := exec.Command("npm", "cache", "clean", "--force")
		// if this fails, we can fallback to rm -rf, but npm should handle it best
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fallback to manual deletion
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	npmDir := filepath.Join(home, ".npm")

	// Check directory exists before removing
	info, err := os.Stat(npmDir)
	if err != nil || !info.IsDir() {
		return nil
	}

	return os.RemoveAll(npmDir)
}
