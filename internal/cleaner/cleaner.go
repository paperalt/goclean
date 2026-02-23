package cleaner

// Cleaner is the interface that all cleaning modules must implement
type Cleaner interface {
	// Name returns the human-readable name of the cleaner
	Name() string
	// Scan returns the amount of bytes that can be cleaned and potential error
	Scan() (int64, error)
	// Clean performs the actual cleanup and returns error if any
	Clean() error
	// RequiresRoot returns true if the cleaner requires root privileges to operate
	RequiresRoot() bool
}
