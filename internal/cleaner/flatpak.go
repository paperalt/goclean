package cleaner

import (
	"fmt"
	"os"
	"os/exec"
)

type FlatpakCleaner struct{}

func (c *FlatpakCleaner) Name() string {
	return "Flatpak (Unused)"
}

func (c *FlatpakCleaner) RequiresRoot() bool {
	// Flatpak commands generally run in user space or ask for polkit auth if needed
	// 'flatpak uninstall --unused' usually runs without sudo for user installs,
	// or might prompt if acting on system level. We mark it false to avoid unnecessary sudo.
	return false
}

func (c *FlatpakCleaner) Scan() (int64, error) {
	// Only scan if flatpak is installed
	if _, err := exec.LookPath("flatpak"); err != nil {
		return 0, nil
	}

	// Flatpak does not provide built-in unused runtime sizes without calculating each one.
	// We'll return 0 to indicate it's a "system tool" rather than a simple folder sweep.
	// Some space will be freed, but we can't reliably predict how much.
	return 0, nil
}

func (c *FlatpakCleaner) Clean() error {
	if _, err := exec.LookPath("flatpak"); err != nil {
		return fmt.Errorf("flatpak not found")
	}

	// We use -y so it auto-confirms
	cmd := exec.Command("flatpak", "uninstall", "--unused", "-y")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
