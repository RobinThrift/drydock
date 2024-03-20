package drydock

import (
	"strings"

	"golang.org/x/exp/slices"
)

func Dir(name string, entries ...File) Directory {
	return &dir{name: name, entries: entries}
}

// DirP is like [Dir] but [name] can be a file path and every
// segment will be created as a directory, similar to [os.Mkdirp] or
// the `mkdir -p` command.
func DirP(name string, entries ...File) Directory {
	dirs := strings.Split(name, "/")
	slices.Reverse(dirs)

	final := Dir(dirs[len(dirs)-1], entries...)

	for _, d := range dirs[:len(dirs)-1] {
		final = Dir(d, final)
	}

	return final
}

type dir struct {
	name    string
	entries []File
}

func (d *dir) Name() string {
	return d.name
}

func (d *dir) Entries() ([]File, error) {
	return d.entries, nil
}
