package drydock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {

	tt := []struct {
		name  string
		input []File
		exp   string
	}{
		{
			name:  "Only Files",
			input: []File{PlainFile("fileA", ""), PlainFile("fileB", ""), PlainFile("fileC", "")},
			exp: `.
├── fileA
├── fileB
└── fileC
`,
		},
		{
			name: "Mixed Files and Nested Dirs",
			input: []File{
				PlainFile("fileA", ""),
				Dir("dirA", PlainFile("dirAFileA", "")),
				Dir("dirB", PlainFile("dirBFileA", ""), PlainFile("dirBFileB", "")),
				PlainFile("fileD", ""),
				Dir("dirC", Dir("dirCdirA",
					PlainFile("dirCdirAFileA", ""),
					Dir("dirCdirADirA", Dir("dirCdirADirADirA",
						PlainFile("dirCdirADirADirAFileA", ""),
						PlainFile("dirCdirADirADirAFileB", ""),
					)),
				)),
			},
			exp: `.
├── fileA
├── dirA/
│   └── dirAFileA
├── dirB/
│   ├── dirBFileA
│   └── dirBFileB
├── fileD
└── dirC/
    └── dirCdirA/
        ├── dirCdirAFileA
        └── dirCdirADirA/
            └── dirCdirADirADirA/
                ├── dirCdirADirADirAFileA
                └── dirCdirADirADirAFileB
`,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			actual := Render(tt.input...)
			assert.Equal(t, tt.exp, actual, actual)
		})
	}
}
