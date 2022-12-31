package swatcher

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	t.Run("valid dependencies", func(t *testing.T) {
		serviceMap := map[string]UserDefinedService{
			"s1": {
				Name:      "s1",
				Cmd:       "echi hi",
				RunBefore: []string{"s2", "s3"},
				Log:       "fd1",
			},
			"s2": {
				Name:      "s2",
				Cmd:       "echo hi2",
				RunBefore: []string{"s3"},
				RunAfter:  []string{"s1"},
				Log:       "fd1",
			},
			"s3": {
				Name:     "s3",
				Cmd:      "echo hi3",
				RunAfter: []string{"s1", "s2"},
				Log:      "fd1",
			},
			"s4": {
				Name:     "s4",
				Cmd:      "echo hi4",
				RunAfter: []string{"s3"},
				Log:      "fd1",
			},
		}

		sortedServices, err := Sort(serviceMap)
		assert.NoError(t, err)

		for i, s1 := range sortedServices {
			for j := i + 1; j < len(sortedServices); j++ {
				s2 := sortedServices[j]
				if strings.Contains(strings.Join(serviceMap[s1.Name].RunAfter, " "), s2.Name) {
					t.Errorf("incorrect ordering. %s should run after %s but it %s came before %s in ordered list", s1.Name, s2.Name, s1.Name, s2.Name)
				}
				if strings.Contains(strings.Join(serviceMap[s2.Name].RunBefore, " "), s1.Name) {
					t.Errorf("incorrect ordering. %s should run before %s, but %s came after %s in ordered list", s2.Name, s1.Name, s2.Name, s1.Name)
				}
			}
		}
	})

	t.Run("dependencies contain a cycle", func(t *testing.T) {
		serviceMap := map[string]UserDefinedService{
			"s1": {
				Name:      "s1",
				Cmd:       "echi hi",
				RunBefore: []string{"s2", "s3"},
				Log:       "fd1",
			},
			"s2": {
				Name:      "s2",
				Cmd:       "echo hi2",
				RunBefore: []string{"s3"},
				RunAfter:  []string{"s1"},
				Log:       "fd1",
			},
			"s3": {
				Name:     "s3",
				Cmd:      "echo hi3",
				RunAfter: []string{"s1", "s2"},
				Log:      "fd1",
			},
			"s4": {
				Name:      "s4",
				Cmd:       "echo hi4",
				RunBefore: []string{"s1"},
				RunAfter:  []string{"s3"},
				Log:       "fd1",
			},
		}

		_, err := Sort(serviceMap)
		assert.Errorf(t, err, "the servcie map contains a cycle. an error should be returned, but was nil")
	})

}
