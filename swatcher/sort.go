package swatcher

// Sort's job is to build relationships between services, and decide which should run before which.
func (m Swatcher) Sort() error
