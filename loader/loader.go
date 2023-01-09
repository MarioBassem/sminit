package loader

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Service struct {
	Name        string
	Cmd         string
	Log         string
	After       []string
	OneShot     bool
	HealthCheck string
}

// LoadAll is responsible for loading all services from /etc/sminit into multiple Service structs
// Users should provide their service definition yaml files in /etc/sminit
// Load ignores any subdirectories in /etc/sminit, or any non-regular files.
func LoadAll(dirPath string) (map[string]Service, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read entries of %s.", dirPath)
	}

	serviceMap := map[string]Service{}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		fullName := entry.Name()
		splitStr := strings.Split(fullName, ".")
		if len(splitStr) != 2 || splitStr[1] != "yaml" {
			return nil, errors.Wrapf(err, "all service definition files should have the \".yaml\" extension. %s is invalid", fullName)
		}

		path := path.Join(dirPath, fullName)
		name := splitStr[0]
		service, err := Load(path, name)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't load service %s.", name)
		}
		serviceMap[service.Name] = service
	}

	return serviceMap, nil
}

// Load is responsible for loading a service from /etc/sminit with a provided serviceName into a Service struct.
func Load(path, serviceName string) (Service, error) {

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
