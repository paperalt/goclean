package cleaner

import (
	"testing"
)

func TestCleanerInterfaces(t *testing.T) {
	cleaners := []Cleaner{
		&AptCleaner{},
		&GoCacheCleaner{},
		&DynamicCacheCleaner{},
		&NpmCacheCleaner{},
		&FlatpakCleaner{},
		&TmpCleaner{},
		&CargoCacheCleaner{},
		&AppCacheCleaner{},
		&BrowserCleaner{},
		&DockerCleaner{},
		&TrashCleaner{},
		&LogCleaner{},
	}

	for _, c := range cleaners {
		name := c.Name()
		if name == "" {
			t.Errorf("Cleaner of type %T has empty name", c)
		}

		// We mostly just check that they don't panic when calling RequiresRoot
		// We shouldn't invoke Clean() or Scan() in a basic unit test automatically
		// because they touch the actual filesystem / root.
		_ = c.RequiresRoot()
	}
}
