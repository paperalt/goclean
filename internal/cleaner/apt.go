package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type AptCleaner struct{}

func (c *AptCleaner) Name() string {
	return "APT Cache"
}

func (c *AptCleaner) RequiresRoot() bool {
	return true
}

func (c *AptCleaner) Scan() (int64, error) {
	// Check if apt-get exists
	_, err := exec.LookPath("apt-get")
	if err != nil {
		return 0, nil // Not an APT system
	}

	// Usually /var/cache/apt/archives
	paths := []string{"/var/cache/apt/archives"}
	var size int64

	for _, p := range paths {
		err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip errors
			}
			if !info.IsDir() && strings.HasSuffix(path, ".deb") {
				size += info.Size()
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	return size, nil
}

func (c *AptCleaner) Clean() error {
	_, err := exec.LookPath("apt-get")
	if err != nil {
		return fmt.Errorf("apt-get not found")
	}

	cmd := exec.Command("sudo", "apt-get", "clean")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
