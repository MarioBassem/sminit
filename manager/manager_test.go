package manager

import (
	"log"
	"testing"
	"time"

	"github.com/mariobassem/sminit-go/loader"
	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	t.Run("simple test", func(t *testing.T) {
		loadedServices := map[string]loader.Service{
			"s1": {
				Name:        "s1",
				Cmd:         "echo mario",
				Log:         "stdout",
				After:       []string{},
				OneShot:     false,
				HealthCheck: "",
			},
			"s2": {
				Name:        "s2",
				Cmd:         "echo bassem",
				Log:         "stdout",
				After:       []string{"s1"},
				OneShot:     false,
				HealthCheck: "",
			},
		}
		manager, err := NewManager(loadedServices)
		assert.NoError(t, err)
		time.Sleep(5 * time.Second)
		log.Print(manager.List())
		addService := loader.Service{
			Name:  "s3",
			Cmd:   "echo gerges",
			Log:   "stdout",
			After: []string{"s1", "s2"},
		}
		err = manager.Add(addService)
		assert.NoError(t, err)
		log.Print(manager.List())
		time.Sleep(2 * time.Second)
		err = manager.Delete("s2")
		assert.NoError(t, err)
		log.Print(manager.List())
		err = manager.Stop("s1")
		assert.NoError(t, err)
		time.Sleep(time.Hour)
	})

}
