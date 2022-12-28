package swatcher

import "github.com/pkg/errors"

// Sort's job is to build relationships between services, and decide which should run before which.
func Sort(services map[string]UserDefinedService) ([]Service, error) {
	// dependencyMap is a map from a service "a" to a list of services that should run after service "a".
	dependencyMap := map[string][]UserDefinedService{}

	for name, service := range services {
		for _, runBeforeStr := range service.RunBefore {
			dependencyMap[name] = append(dependencyMap[name], services[runBeforeStr])
		}

		for _, runAfterStr := range service.RunAfter {
			dependencyMap[runAfterStr] = append(dependencyMap[runAfterStr], services[name])
		}
	}

	sortedServices, err := topologicalSort(dependencyMap)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't sort services.")
	}

	return sortedServices, nil
}

// topologicalSort should sort the services topologically. it should return an error if the dependecyMap contains a cycle.
func topologicalSort(dependencyMap map[string]UserDefinedService) ([]Service, error) {

}
