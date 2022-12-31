package swatcher

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// LoadAll is responsible for loading all services from /etc/sminit into multiple Service structs
// Users should provide their service definition yaml files in /etc/sminit
// Load ignores any subdirectories in /etc/sminit, or any non-regular files.
func (s Swatcher) LoadAll(dirPath string) ([]Service, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read entries of %s.", dirPath)
	}

	services := []Service{}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		name := entry.Name()
		service, err := s.Load(dirPath, name)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't load service %s.", name)
		}

		services = append(services, service)
	}

	return services, nil
}

// Load is responsible for loading a service from /etc/sminit with a provided serviceName into a Service struct.
func (s Swatcher) Load(dirPath, serviceName string) (Service, error) {
	path := path.Join(dirPath, serviceName)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return Service{}, errors.Wrapf(err, "couldn't read contents of file %s.", path)
	}

	service := Service{
		Name: serviceName,
	}

	err = yaml.Unmarshal(bytes, &service)
	if err != nil {
		return Service{}, errors.Wrapf(err, "couldn't unmarshal contents of file %s.", path)
	}

	return service, nil
}
