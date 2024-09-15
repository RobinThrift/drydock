package drydock

import (
	"io/fs"
)

// OutputFS extends the standard [io/fs.FS] interface with writing capabilities for
// hierarchical file systems.
type OutputFS interface {
	fs.FS

	OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error)

	// Mkdir creates a new directory and must return fs.ErrExist if the directory already exists.
	Mkdir(name string) error

	// Rename a given file or directory. Returned errors must be [os.LinkError] errors.
	Rename(oldpath string, newpath string) error

	// Remove the file at path or directory, if the directory is empty.
	Remove(path string) error

	// RemoveAll is like [OutputFS.Remove] but also removes any non-empty directories.
	RemoveAll(path string) error

	// MkdirTemp creates a temporary directory with the pattern, as described in [os.MkdirTemp].
	// The caller is responsible for removing and temporary directories, they will not be cleaned up automatically.
	MkdirTemp(pattern string) (OutputFS, string, error)
}
