package swatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwatch(t *testing.T) {
	t.Run("simple test", func(t *testing.T) {
		err := Swatch()
		assert.NoError(t, err)
	})
}
