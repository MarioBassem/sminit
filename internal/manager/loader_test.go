package manager

import (
	"bytes"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLoader(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("load_all", func(t *testing.T) {
		want := map[string]ServiceOptions{
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
		want := map[string]ServiceOptions{
			"s1": {
				Name:        "s1",
				Cmd:         "echo hi",
				After:       []string{"s2", "s3"},
				OneShot:     true,
				HealthCheck: "sleep 5",
				Log:         "log1",
			},
		}

		wantBytes, err := yaml.Marshal(want["s1"])
		assert.NoError(t, err)

		serviceOptions, err := ReadService(bytes.NewReader(wantBytes), "s1")
		assert.NoError(t, err)
		assert.Equal(t, want["s1"], serviceOptions)
	})
}

func WriteServices(dir string, serviceOptionsMap map[string]ServiceOptions) error {
	for _, opt := range serviceOptionsMap {
		path := path.Join(dir, strings.Join([]string{opt.Name, ".yaml"}, ""))
		file, err := os.Create(path)
		if err != nil {
			return errors.Wrapf(err, "could not create definition file for service %s", opt.Name)
		}

		bytes, err := yaml.Marshal(opt)
		if err != nil {
			return errors.Wrapf(err, "could not marshal service %s to bytes", opt.Name)
		}

		_, err = file.Write(bytes)
		if err != nil {
			return err
		}
	}
	return nil
}
