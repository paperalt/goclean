package cleaner

import (
	"os"
	"path/filepath"
)

type CargoCacheCleaner struct{}

func (c *CargoCacheCleaner) Name() string {
	return "Cargo Cache (Rust)"
}

func (c *CargoCacheCleaner) RequiresRoot() bool {
	return false
}

func (c *CargoCacheCleaner) Scan() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	var size int64
	paths := []string{
		filepath.Join(home, ".cargo", "registry", "cache"),
		filepath.Join(home, ".cargo", "registry", "src"),
		filepath.Join(home, ".cargo", "git", "db"),
		filepath.Join(home, ".cargo", "git", "checkouts"),
	}

	for _, p := range paths {
		s, _ := simpleDirScan(p)
		size += s
	}

	return size, nil
}

func (c *CargoCacheCleaner) Clean() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	paths := []string{
		filepath.Join(home, ".cargo", "registry", "cache"),
		filepath.Join(home, ".cargo", "registry", "src"),
		filepath.Join(home, ".cargo", "git", "db"),
		filepath.Join(home, ".cargo", "git", "checkouts"),
	}

	for _, p := range paths {
		// Just blow away the cache folders directly. Cargo will re-download on demand.
		_ = os.RemoveAll(p)
	}

	return nil
}
