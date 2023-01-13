package swatch

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ServiceOptions struct {
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
func LoadAll(servicesDirPath string) (map[string]ServiceOptions, error) {
	entries, err := os.ReadDir(servicesDirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read entries of %s.", servicesDirPath)
	}

	optionsMap := map[string]ServiceOptions{}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		serviceOptionsFileName := entry.Name()
		splittedFileName := strings.Split(serviceOptionsFileName, ".")
		if len(splittedFileName) != 2 || splittedFileName[1] != "yaml" {
			return nil, errors.Wrapf(err, `%s does not have the .yaml extension`, serviceOptionsFileName)
		}

		path := path.Join(servicesDirPath, serviceOptionsFileName)
		name := splittedFileName[0]
		service, err := Load(path, name)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't load service %s.", name)
		}
		optionsMap[service.Name] = service
	}

	return optionsMap, nil
}

// Load is responsible for loading a service from /etc/sminit with a provided serviceName into a Service struct.
func Load(path, serviceName string) (ServiceOptions, error) {

	bytes, err := os.ReadFile(path)
	if err != nil {
		return ServiceOptions{}, errors.Wrapf(err, "couldn't read contents of file %s.", path)
	}

	service := ServiceOptions{
		Name: serviceName,
	}

	err = yaml.Unmarshal(bytes, &service)
	if err != nil {
		return ServiceOptions{}, errors.Wrapf(err, "couldn't unmarshal contents of file %s.", path)
	}

	return service, nil
}
