package bt

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestParseLinkKey(t *testing.T) {
	tests := []struct {
		input    string
		expected LinkKey
		hasError bool
	}{
		{"9EF14F8CB54D8B01048F4A8F4A8F4A8F", LinkKey("9EF14F8CB54D8B01048F4A8F4A8F4A8F"), false},
		{"hex:c5,cc,96,ec,48,ee,88,8f,04,a8,63,34,4c,c6,a7,2d", LinkKey("C5CC96EC48EE888F04A863344CC6A72D"), false},
		{"invalid_key", "", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseLinkKey(test.input)
			if test.hasError {
				assert.ErrorContains(t, err, "invalid link key format")
			} else {
				assert.NilError(t, err)
				assert.Check(t, cmp.Equal(result, test.expected))
			}
		})
	}
}
