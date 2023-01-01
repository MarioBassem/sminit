package swatcher

import (
	"fmt"

	"github.com/pkg/errors"
)

// validateServiceDependency checks if there is a non-existent dependency in each service's dependency lists.
func validateServiceDependency(services []Service) error {
	existentServices := map[string]bool{}
	for _, service := range services {
		existentServices[service.Name] = true
	}

	for _, service := range services {
		for _, serviceName := range service.RunAfter {
			if _, ok := existentServices[serviceName]; !ok {
				return fmt.Errorf("service %s does not exist", serviceName)
			}
		}

		for _, serviceName := range service.RunBefore {
			if _, ok := existentServices[serviceName]; !ok {
				return fmt.Errorf("service %s does not exist", serviceName)
			}
		}
	}

	return nil
}

// validateServiceLog validates the log path provided in the service definition file
func validateServiceLog(services []Service) error {
	for _, service := range services {
		if service.Log != "" && service.Log != "stdout" {
			return fmt.Errorf("service %s could only be empty or \"stdout\"", service.Name)
		}
	}
	return nil
}

// Validate validates services' dependencies and logs, returns an error if either is invalid.
func (s Swatcher) Validate(services []Service) error {
	err := validateServiceDependency(services)
	if err != nil {
		return errors.Wrap(err, "couldn't validate services' dependencies")
	}

	err = validateServiceLog(services)
	if err != nil {
		return errors.Wrap(err, "couldn't validate services' logs")
	}
	return nil
}
