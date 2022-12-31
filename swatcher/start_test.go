package swatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	t.Run("start", func(t *testing.T) {
		sortedServices := []Service{
			{
				Name:   "s1",
				CmdStr: "echo \"service s1 hiii\"",
			},
			{
				Name:   "s2",
				CmdStr: "echo \"service s2 hiii\"",
			},
		}
		swatcher, err := NewSwatcher()
		swatcher.SortedServices = sortedServices
		assert.NoError(t, err)

		err = swatcher.Start()
		assert.NoError(t, err)
	})
}
