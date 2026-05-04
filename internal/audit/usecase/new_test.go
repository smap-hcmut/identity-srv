package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCleanupJob(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  struct{}
		mock   mock
		output bool
		err    error
	}{
		"success": {
			output: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := NewCleanupJob(nil, nil)

			require.NotNil(t, output)
			assert.Equal(t, tc.output, output.cron != nil)
			assert.NoError(t, tc.err)
		})
	}
}
