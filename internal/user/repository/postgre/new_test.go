package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
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
			output := New(nil, nil)

			require.NotNil(t, output)
			assert.Equal(t, tc.output, output.clock != nil)
			assert.NoError(t, tc.err)
		})
	}
}
