package loader

import (
	"os"
	"path"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLoader(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("load all", func(t *testing.T) {
		want := map[string]Service{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hi",
				After:       []string{"s2", "s3"},
				OneShot:     true,
				HealthCheck: "sleep 5",
				Log:         "log1",
			},
			"s2": {
				Name:        "s2",
				Cmd:         "echo hi",
				After:       []string{"s3"},
				OneShot:     true,
				HealthCheck: "sleep 5",
				Log:         "log2",
			},
			"s3": {
				Name:    "s3",
				Cmd:     "echo hi",
				After:   []string{"s2", "s3"},
				OneShot: true,
				Log:     "log1",
			},
			"s4": {
				Name:        "s4",
				Cmd:         "echo hi",
				After:       []string{"s2", "s3"},
				OneShot:     false,
				HealthCheck: "sleep 5",
				Log:         "log1",
			},
		}
		err := WriteServices(tmpDir, want)
		assert.NoError(t, err)

		services, err := LoadAll(tmpDir)
		assert.NoError(t, err)
		assert.Equal(t, want, services)
	})

	t.Run("load one", func(t *testing.T) {
		want := map[string]Service{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hi",
				After:       []string{"s2", "s3"},
				OneShot:     true,
				HealthCheck: "sleep 5",
				Log:         "log1",
			},
		}

		err := WriteServices(tmpDir, want)
		assert.NoError(t, err)

		service, err := Load(tmpDir, "s1")
		assert.NoError(t, err)
		assert.Equal(t, want["s1"], service)
	})
}

func WriteServices(dir string, services map[string]Service) error {
	for _, service := range services {
		path := path.Join(dir, service.Name)
		file, err := os.Create(path)
		if err != nil {
			return errors.Wrapf(err, "couldn't create definition file for service %s", service.Name)
		}

		bytes, err := yaml.Marshal(service)
		if err != nil {
			return errors.Wrapf(err, "couldn't marshal service %s to bytes", service.Name)
		}

		_, err = file.Write(bytes)
		if err != nil {
			return errors.Wrap(err, "couldn't write bytes")
		}
	}
	return nil
}
