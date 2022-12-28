package swatcher

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// LoadAll is responsible for loading all services from /etc/sminit into multiple UserDefinedService structs
// Users should provide their service definition yaml files in /etc/sminit
// Load ignores any subdirectories in /etc/sminit, or any non-regular files.
func (s Swatcher) LoadAll(dirPath string) ([]UserDefinedService, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read entries of %s.", dirPath)
	}

	services := []UserDefinedService{}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		name := entry.Name()
		userDefinedService, err := s.Load(dirPath, name)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't load service %s.", name)
		}

		services = append(services, userDefinedService)
	}

	return services, nil
}

// Load is responsible for loading a service from /etc/sminit with a provided serviceName into a UserDefinedService struct.
func (s Swatcher) Load(dirPath, serviceName string) (UserDefinedService, error) {
	path := path.Join(dirPath, serviceName)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return UserDefinedService{}, errors.Wrapf(err, "couldn't read contents of file %s.", path)
	}

	userDefinedServide := UserDefinedService{
		Name: serviceName,
	}

	err = yaml.Unmarshal(bytes, &userDefinedServide)
	if err != nil {
		return UserDefinedService{}, errors.Wrapf(err, "couldn't unmarshal contents of file %s.", path)
	}

	return userDefinedServide, nil
}
