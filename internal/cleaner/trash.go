package cleaner

import (
	"os"
	"path/filepath"
)

type TrashCleaner struct{}

func (c *TrashCleaner) Name() string {
	return "User Trash"
}

func (c *TrashCleaner) RequiresRoot() bool {
	return false
}

func (c *TrashCleaner) Scan() (int64, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}
	
	trashPath := filepath.Join(home, ".local", "share", "Trash")
	var size int64
	
	err = filepath.Walk(trashPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, nil
}

func (c *TrashCleaner) Clean() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	trashPath := filepath.Join(home, ".local", "share", "Trash")
	
	// We destroy everything inside files/ and info/ directories usually, 
	// but nuking the whole Trash folder content is also acceptable for "Clean Trash".
	// Let's just remove the folder contents to be clean.
	return os.RemoveAll(trashPath)
}
