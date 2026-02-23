package cleaner

import (
	"fmt"
	"os/exec"
	"strings"
)

type GoCacheCleaner struct{}

func (c *GoCacheCleaner) Name() string {
	return "Go Build Cache"
}

func (c *GoCacheCleaner) RequiresRoot() bool {
	return false
}

func (c *GoCacheCleaner) Scan() (int64, error) {
	if _, err := exec.LookPath("go"); err != nil {
		return 0, nil
	}

	// `go clean -cache -n` prints what it would remove.
	// But listing the directory size of `go env GOCACHE` is better.
	
	cmd := exec.Command("go", "env", "GOCACHE")
	output, err := cmd.Output()
	if err != nil {
		return 0, nil
	}
	
	cachePath := strings.TrimSpace(string(output))
	if cachePath == "" {
		return 0, nil
	}
	
	// Scan dir
	return simpleDirScan(cachePath)
}

func (c *GoCacheCleaner) Clean() error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found")
	}
	return exec.Command("go", "clean", "-cache").Run()
}
