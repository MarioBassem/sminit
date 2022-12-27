package swatcher

// load is responsible for loading all services from /etc/sminit into multiple `Service` structs
func (m Swatcher) Load() error
