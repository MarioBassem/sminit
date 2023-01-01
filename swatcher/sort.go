package swatcher

import (
	"github.com/mariobassem/sminit-go/stack"
	"github.com/pkg/errors"
)

// Sort's job is to build relationships between services, and decide which should run before which.
func Sort(serviceMap map[string]Service) ([]Service, error) {
	// dependencyMap is a map from a service "a" to a list of services that should run after service "a".
	dependencyMap := map[string][]string{}

	for name, service := range serviceMap {
		dependencyMap[name] = append(dependencyMap[name], service.RunBefore...)

		for _, runAfterStr := range service.RunAfter {
			dependencyMap[runAfterStr] = append(dependencyMap[runAfterStr], name)
		}
	}

	sortedServices, err := topologicalSort(dependencyMap, serviceMap)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't sort services.")
	}

	return sortedServices, nil
}

// topologicalSort should sort the services topologically. it should return an error if the dependecyMap contains a cycle.
func topologicalSort(dependencyMap map[string][]string, serviceMap map[string]Service) ([]Service, error) {
	vis := map[string]int{}
	st := stack.Stack[string]{}
	for service := range dependencyMap {
		if vis[service] == 2 {
			continue
		}
		err := dfs(service, dependencyMap, vis, &st)
		if err != nil {
			return nil, err
		}
	}

	sortedServices := []Service{}
	for !st.Empty() {
		top, err := st.Top()
		if err != nil {
			// this shouldn't happen because Top() returns an error if stack is empty, which we ensured is not true.
			return nil, err
		}
		sortedServices = append(sortedServices, serviceMap[top])
	}

	return sortedServices, nil
}

func dfs(currentServiceName string, dependencyMap map[string][]string, vis map[string]int, st *stack.Stack[string]) error {
	vis[currentServiceName] = 1
	for _, dependentService := range dependencyMap[currentServiceName] {
		if vis[dependentService] == 1 {
			// this is a service that is still being processed: this is a cycle, and an error should be returned.
			return errors.Errorf("found a cycle, service %s shouldn't depend on service %s.", dependentService, currentServiceName)
		}

		if vis[dependentService] == 2 {
			// this is an already processed service. it should be ignored.
			continue
		}

		err := dfs(dependentService, dependencyMap, vis, st)
		if err != nil {
			return err
		}
	}

	vis[currentServiceName] = 2
	st.Push(currentServiceName)

	return nil
}
