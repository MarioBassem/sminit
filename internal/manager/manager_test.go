package manager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	t.Run("simple_test", func(t *testing.T) {
		loadedServices := map[string]ServiceOptions{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hello",
				Log:         "stdout",
				After:       []string{},
				OneShot:     false,
				HealthCheck: "",
			},
		}
		manager, err := NewManager(loadedServices)
		assert.NoError(t, err)

		manager.fireServices()

		listServices := manager.List()
		assert.True(t, listServices[0].Name == "s1")
	})

	t.Run("dependency_test", func(t *testing.T) {
		loadedServices := map[string]ServiceOptions{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hello",
				Log:         "stdout",
				After:       []string{},
				OneShot:     false,
				HealthCheck: "sleep 2",
			},
			"s2": {
				Name:        "s2",
				Cmd:         "echo world",
				Log:         "stdout",
				After:       []string{"s1"},
				OneShot:     false,
				HealthCheck: "",
			},
		}
		manager, err := NewManager(loadedServices)
		assert.NoError(t, err)

		manager.fireServices()

		time.Sleep(time.Second)

		listServices := manager.List()
		var s2 Service
		if listServices[0].Name == "s2" {
			s2 = listServices[0]
		} else {
			s2 = listServices[1]
		}
		assert.True(t, s2.Status == Pending)

		time.Sleep(2 * time.Second)

		listServices = manager.List()
		if listServices[0].Name == "s2" {
			s2 = listServices[0]
		} else {
			s2 = listServices[1]
		}
		assert.True(t, s2.Status != Pending)
	})

	t.Run("delete_test", func(t *testing.T) {
		loadedServices := map[string]ServiceOptions{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hello",
				Log:         "stdout",
				After:       []string{},
				OneShot:     false,
				HealthCheck: "",
			},
			"s2": {
				Name:        "s2",
				Cmd:         "echo world",
				Log:         "stdout",
				After:       []string{"s1"},
				OneShot:     false,
				HealthCheck: "",
			},
		}
		manager, err := NewManager(loadedServices)
		assert.NoError(t, err)

		manager.fireServices()

		err = manager.Delete("s2")
		assert.NoError(t, err)

		listServices := manager.List()
		assert.True(t, listServices[0].Name == "s1")
		assert.True(t, len(listServices) == 1)
	})

	t.Run("stop_test", func(t *testing.T) {
		loadedServices := map[string]ServiceOptions{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hello",
				Log:         "stdout",
				After:       []string{},
				OneShot:     false,
				HealthCheck: "",
			},
		}
		manager, err := NewManager(loadedServices)
		assert.NoError(t, err)

		manager.fireServices()

		err = manager.Stop("s1")
		assert.NoError(t, err)

		time.Sleep(time.Second)

		list := manager.List()
		assert.True(t, list[0].Status == Stopped, "status is %s", list[0].Status)
	})

	t.Run("start_test", func(t *testing.T) {
		loadedServices := map[string]ServiceOptions{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hello",
				Log:         "stdout",
				After:       []string{},
				OneShot:     false,
				HealthCheck: "",
			},
		}
		manager, err := NewManager(loadedServices)
		assert.NoError(t, err)

		manager.fireServices()

		err = manager.Stop("s1")
		assert.NoError(t, err)

		err = manager.Start("s1")
		assert.NoError(t, err)

		list := manager.List()
		assert.True(t, list[0].Status != Stopped)
	})

}
