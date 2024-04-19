package drydock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirP(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		entries []File
		exp     Directory
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
		{
			name:    "Nested Dir With Files",
			input:   "Level-0/Level-1/Level-2",
			entries: []File{PlainFile("fileA", "fileA contents")},
			exp:     Dir("Level-0", Dir("Level-1", Dir("Level-2", PlainFile("fileA", "fileA contents")))),
		},
		{
			name:    "Dot Directory",
			input:   "./Level-0/Level-1",
			entries: []File{PlainFile("FileNameA", "")},
			exp:     Dir("Level-0", Dir("Level-1", PlainFile("FileNameA", ""))),
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			actual := DirP(tt.input, tt.entries...)
			assert.Equal(t, tt.exp, actual)
		})
	}
}
