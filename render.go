package drydock

import (
	"strings"
)

// Render returns a textual representation of the file structure, primarily used for debugging.
func Render(files ...File) string {
	var b strings.Builder

	var dir Directory

	if len(files) == 1 {
		if d, isDir := files[0].(Directory); isDir {
			dir = d
		}
	} else {
		dir = Dir(".", files...)
	}

	b.WriteString(dir.Name() + "\n")

	entries, _ := dir.Entries()

	for i, entry := range entries {
		isLast := i == len(entries)-1
		if dir, ok := entry.(Directory); ok {
			renderDir(&b, dir, 0, isLast)
		} else {
			prefix(&b, 0, false, isLast)
			b.WriteString(entry.Name() + "\n")
		}
	}

	return b.String()
}

func renderDir(b *strings.Builder, dir Directory, level int, isLast bool) {
	prefix(b, level, isLast, isLast)
	b.WriteString(dir.Name() + "/\n")

	entries, _ := dir.Entries()

	for i, entry := range entries {
		lastEntry := i == len(entries)-1
		if dir, ok := entry.(Directory); ok {
			renderDir(b, dir, level+1, lastEntry)
		} else {
			prefix(b, level+1, isLast, lastEntry)
			b.WriteString(entry.Name() + "\n")
		}
	}
}

func prefix(b *strings.Builder, level int, parentIsLast bool, isLast bool) {
	for i := 0; i < level; i++ {
		if parentIsLast {
			b.WriteString("    ")
		} else {
			b.WriteString("│   ")
		}
	}

	if isLast {
		b.WriteString("└── ")
	} else {
		b.WriteString("├── ")
	}
}
