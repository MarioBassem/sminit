package swatcher

// Wait's job is to wait for requests from other processes and relay them to the manager
func (m Swatcher) Wait() error
