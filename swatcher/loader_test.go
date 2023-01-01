package swatcher

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
		want := []Service{
			{
				Name:      "s1",
				Cmd:       "echo hi",
				RunBefore: []string{"s2", "s3"},
				RunAfter:  []string{"s4"},
				Log:       "log1",
			},
			{
				Name:      "s2",
				Cmd:       "echo \"hello world\"",
				RunBefore: []string{"s1", "s3"},
				RunAfter:  []string{"s2", "s4"},
			},
			{
				Name:      "s3",
				Cmd:       "echo \"hello world 3\"",
				RunBefore: []string{"s1", "s3"},
				RunAfter:  []string{"s2", "s4"},
			},
			{
				Name:      "s4",
				Cmd:       "echo \"hello world\"",
				RunBefore: []string{"s1", "s3"},
				RunAfter:  []string{"s2", "s4"},
			},
		}

		err := WriteServices(tmpDir, want)
		assert.NoError(t, err)

		swathcer, err := NewSwatcher()
		assert.NoError(t, err)

		services, err := swathcer.LoadAll(tmpDir)
		assert.NoError(t, err)
		assert.Equal(t, want, services)
	})

	t.Run("load one", func(t *testing.T) {
		want := Service{
			Name:      "s1",
			Cmd:       "echo hi",
			RunBefore: []string{"s2", "s3"},
			RunAfter:  []string{"s4"},
			Log:       "log1",
		}

		err := WriteServices(tmpDir, []Service{want})
		assert.NoError(t, err)

		swatcher, err := NewSwatcher()
		assert.NoError(t, err)

		service, err := swatcher.Load(tmpDir, want.Name)
		assert.NoError(t, err)
		assert.Equal(t, want, service)
	})

	t.Run("non-existent dependency", func(t *testing.T) {
		service := Service{
			Name:      "s1",
			Cmd:       "echo hi",
			RunBefore: []string{"s2"},
			RunAfter:  []string{"s1"},
			Log:       "log1",
		}
		err := WriteServices(tmpDir, []Service{service})
		assert.NoError(t, err)

		swatcher, err := NewSwatcher()
		assert.NoError(t, err)

		_, err = swatcher.Load(tmpDir, service.Name)
		assert.Error(t, err, "loader should have returned an error due to non-existent dependency, but returned nil")

	})
}

func WriteServices(dir string, services []Service) error {
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
