package drydock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirP(t *testing.T) {
	tt := []struct {
		name  string
		input string
		exp   Directory
	}{
		{
			name:  "Single Directory",
			input: "DirName",
			exp:   Dir("DirName"),
		},
		{
			name:  "Nested Dir",
			input: "Level-0/Level-1/Level-2",
			exp:   Dir("Level-0", Dir("Level-1", Dir("Level-2"))),
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			actual := DirP(tt.input)
			assert.Equal(t, tt.exp, actual)
		})
	}
}
