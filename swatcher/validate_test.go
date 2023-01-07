package swatcher

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestValidate(t *testing.T) {
// 	t.Run("valid services", func(t *testing.T) {
// 		services := []Service{
// 			{
// 				Name:      "s1",
// 				RunBefore: []string{"s2", "s3"},
// 			},
// 			{
// 				Name:      "s2",
// 				RunBefore: []string{"s3"},
// 				RunAfter:  []string{"s1"},
// 				Log:       "stdout",
// 			},
// 			{
// 				Name:     "s3",
// 				RunAfter: []string{"s1", "s2"},
// 			},
// 		}
// 		swatcher, err := NewSwatcher()
// 		assert.NoError(t, err)

// 		err = swatcher.Validate(services)
// 		assert.NoError(t, err)
// 	})

// 	t.Run("non-existent dependency", func(t *testing.T) {
// 		services := []Service{
// 			{
// 				Name:      "s1",
// 				RunBefore: []string{"s2", "s3"},
// 			},
// 			{
// 				Name:      "s2",
// 				RunBefore: []string{"s3"},
// 				RunAfter:  []string{"s1", "s4"},
// 				Log:       "stdout",
// 			},
// 			{
// 				Name:     "s3",
// 				RunAfter: []string{"s1", "s2"},
// 			},
// 		}
// 		swatcher, err := NewSwatcher()
// 		assert.NoError(t, err)

// 		err = swatcher.Validate(services)
// 		assert.Error(t, err)
// 	})

// 	t.Run("invalid log path", func(t *testing.T) {
// 		services := []Service{
// 			{
// 				Name:      "s1",
// 				RunBefore: []string{"s2", "s3"},
// 			},
// 			{
// 				Name:      "s2",
// 				RunBefore: []string{"s3"},
// 				RunAfter:  []string{"s1"},
// 				Log:       "stdout",
// 			},
// 			{
// 				Name:     "s3",
// 				RunAfter: []string{"s1", "s2"},
// 				Log:      "stderr",
// 			},
// 		}

// 		swatcher, err := NewSwatcher()
// 		assert.NoError(t, err)

// 		err = swatcher.Validate(services)
// 		assert.Error(t, err)
// 	})
// }
