package swatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	t.Run("start", func(t *testing.T) {
		sortedServices := []Service{
			{
				Name: "s1",
				Cmd:  "ping google.com",
				Log:  "stdout",
			},
			{
				Name: "s2",
				Cmd:  "echo \"service s2 hiii\"",
				Log:  "stdout",
			},
		}
		swatcher, err := NewSwatcher()
		swatcher.SortedServices = sortedServices
		assert.NoError(t, err)

		err = swatcher.Start()
		assert.NoError(t, err)
	})
}
