package drydock

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

// WritableFS extends the standard [io/fs.FS] interface with writing capabilities for
// hierarchical file systems.
type WritableFS interface {
	fs.ReadFileFS

	// Mkdir should behave like [os.Mkdir].
	Mkdir(name string, perm fs.FileMode) error

	// Rename should behave like [os.Rename]. Returned errors will also be [os.LinkError].
	Rename(oldpath string, newpath string) error

	// Remove should behave like [os.Remove].
	Remove(path string) error

	// RemoveAll should behave like [os.RemoveAll].
	RemoveAll(path string) error

	// CreateTemp should behave like [os.CreateTemp].
	CreateTemp(dir string, pattern string) (WritableFile, error)
}

type WritableFile interface {
	fs.File
	io.Writer
	Name() string
}

type writableDirFS struct {
	fs.ReadFileFS
	baseDir string
}

// NewWritableDirFS creates a new [WritablesFS] backed by the real filesystem,
// like [os.DirFS].
func NewWritableDirFS(dir string) WritableFS {
	return &writableDirFS{ReadFileFS: os.DirFS(dir).(fs.ReadFileFS), baseDir: dir}
}

// MkWritableDirFS is like [NewWritableDirFS] but will create the directory if it doesn't exist.
func MkWritableDirFS(dir string) (WritableFS, error) {
	err := os.Mkdir(dir, 0755)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}

	return NewWritableDirFS(dir), nil
}

func (wfs *writableDirFS) Mkdir(name string, perm fs.FileMode) error {
	err := os.Mkdir(path.Join(wfs.baseDir, name), perm)
	return err
}

func (wfs *writableDirFS) Rename(oldpath string, newpath string) error {
	if !strings.HasPrefix(newpath, wfs.baseDir) {
		newpath = path.Join(wfs.baseDir, newpath)
	}

	return os.Rename(oldpath, newpath)
}

func (wfs *writableDirFS) Remove(p string) error {
	return os.Remove(path.Join(wfs.baseDir, p))
}

func (wfs *writableDirFS) RemoveAll(p string) error {
	return os.RemoveAll(path.Join(wfs.baseDir, p))
}

func (wfs *writableDirFS) CreateTemp(dir string, pattern string) (WritableFile, error) {
	return os.CreateTemp(dir, pattern)
}

func (wfs *writableDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return wfs.ReadFileFS.(fs.ReadDirFS).ReadDir(name)
}
