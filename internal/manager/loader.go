package manager

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ServiceOptions struct {
	Name        string `yaml:"omitempty"`
	Cmd         string
	Log         string
	After       []string
	OneShot     bool
	HealthCheck string
}

// LoadAll is responsible for loading all services from /etc/sminit into multiple Service structs
// Users should provide their service definition yaml files in /etc/sminit
func LoadAll(servicesDirPath string) (map[string]ServiceOptions, error) {
	entries, err := os.ReadDir(servicesDirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read entries of %s", servicesDirPath)
	}

	optionsMap := make(map[string]ServiceOptions)

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			return nil, fmt.Errorf("all entries in %s should be regular files", servicesDirPath)
		}

		serviceOptionsFileName := entry.Name()

		splittedFileName := strings.Split(serviceOptionsFileName, ".")
		if len(splittedFileName) != 2 || splittedFileName[1] != "yaml" {
			return nil, errors.Wrapf(err, `%s does not have .yaml extension`, serviceOptionsFileName)
		}

		path := path.Join(servicesDirPath, serviceOptionsFileName)

		name := splittedFileName[0]

		file, err := os.Open(path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not open file at %s", path)
		}

		service, err := ServiceReader(file, name)
		if err != nil {
			return nil, errors.Wrapf(err, "could not load service %s.", name)
		}

		optionsMap[service.Name] = service
	}

	return optionsMap, nil
}

// ServiceReader is responsible for loading a service from /etc/sminit with a provided serviceName into a Service struct.
func ServiceReader(reader io.Reader, serviceName string) (ServiceOptions, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return ServiceOptions{}, errors.Wrapf(err, "could not read service %s options from reader", serviceName)
	}

	service := ServiceOptions{
		Name: serviceName,
	}

	err = yaml.Unmarshal(bytes, &service)
	if err != nil {
		return ServiceOptions{}, errors.Wrapf(err, "could not unmarshal bytes contents %s", (bytes))
	}

	return service, nil
}
