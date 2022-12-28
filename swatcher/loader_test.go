package swatcher

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLoader(t *testing.T) {
	tmpDir := t.TempDir()
	want := []UserDefinedService{
		{
			Name:      "s1",
			Cmd:       "echo hi",
			RunBefore: []string{"s2", "s3"},
			RunAfter:  []string{"s4", "s5"},
			Log:       "log1",
		},
		{
			Name:      "s2",
			Cmd:       "echo \"hello world\"",
			RunBefore: []string{"s1", "s3"},
			RunAfter:  []string{"s2", "s4"},
		},
	}

	for _, service := range want {
		path := path.Join(tmpDir, service.Name)
		file, err := os.Create(path)
		if err != nil {
			t.Errorf("couldn't create definition file for service %s. %s", service.Name, err.Error())
		}

		bytes, err := yaml.Marshal(service)
		if err != nil {
			t.Errorf("couldn't marshal service %s to bytes. %s", service.Name, err.Error())
		}

		_, err = file.Write(bytes)
		if err != nil {
			t.Errorf("couldn't write bytes. %s", err.Error())
		}
	}
	t.Run("load all", func(t *testing.T) {
		swathcer, err := NewSwatcher()
		assert.NoError(t, err)

		services, err := swathcer.LoadAll(tmpDir)
		assert.NoError(t, err)
		assert.Equal(t, want, services)
	})

	t.Run("load one", func(t *testing.T) {
		swatcher, err := NewSwatcher()
		assert.NoError(t, err)

		service, err := swatcher.Load(tmpDir, want[0].Name)
		assert.NoError(t, err)
		assert.Equal(t, want[0], service)
	})
}
