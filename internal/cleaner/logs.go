package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type LogCleaner struct{}

func (c *LogCleaner) Name() string {
	return "System Logs"
}

func (c *LogCleaner) RequiresRoot() bool {
	return true
}

func (c *LogCleaner) Scan() (int64, error) {
	var size int64
	
	// 1. Scan /var/log for rotated logs (*.gz, *.[0-9])
	err := filepath.Walk("/var/log", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".1") || strings.HasSuffix(path, ".old") {
				size += info.Size()
			}
		}
		return nil
	})
	if err != nil {
		// ignore perm errors during scan for now, or maybe return?
	}

	// 2. Journald vacuum size estimation is hard without running it, 
	// but we can check check disk usage of /var/log/journal if it exists
	jPath := "/var/log/journal"
	if _, err := os.Stat(jPath); err == nil {
		// Just count everything in journal? No, that's dangerous. 
		// Let's rely on vacuum-time=3d for cleaning, but for scanning...
		// Maybe just skip scanning journald specific bytes for now to be safe and accurate on "reclaimable"
	}

	return size, nil
}

func (c *LogCleaner) Clean() error {
	// vacuum systemd journal
	if _, err := exec.LookPath("journalctl"); err == nil {
		cmd := exec.Command("sudo", "journalctl", "--vacuum-time=3d")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error vacuuming journal: %v\n", err)
		}
	}

	// Delete rotated logs in /var/log
	return filepath.Walk("/var/log", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".1") || strings.HasSuffix(path, ".old") {
				// We need sudo to remove likely
				// But we are running the binary. If binary run as sudo, os.Remove works.
				// If not, we might fail.
				// For this robust implementation, let's assume user runs with sudo for deep clean.
				// Or we shell out to sudo rm? simpler to try os.Remove and report error if perm denied.
				if err := os.Remove(path); err != nil {
					// try to run sudo rm if permission denied? 
					// simpler: just exec sudo rm
					exec.Command("sudo", "rm", path).Run()
				}
			}
		}
		return nil
	})
}
