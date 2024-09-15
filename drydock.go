package drydock

import (
	"io"
)

// File is either a real file to be generated or a directory.
// Plain files must also implement [io.Writer] to write their content to the output file.
type File interface {
	Name() string
}

// A Directory can contain files and other directories.
type Directory interface {
	File
	Entries() ([]File, error)
}

type WriterToModify interface {
	WriteModifiedTo(contents []byte, w io.Writer) error
}
