package drydock

import (
	"io/fs"
	"os"
	"path"
	"strings"
)

type osOutputFS struct {
	fs.ReadFileFS
	baseDir string
}

// NewOSOutputFS creates a new [OutputsFS] backed by the real filesystem,
// like [os.DirFS].
func NewOSOutputFS(dir string) OutputFS {
	return &osOutputFS{ReadFileFS: os.DirFS(dir).(fs.ReadFileFS), baseDir: dir}
}

func (ofs *osOutputFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	return os.OpenFile(path.Join(ofs.baseDir, name), flag, perm)
}

func (ofs *osOutputFS) Mkdir(name string) error {
	err := os.Mkdir(path.Join(ofs.baseDir, name), 0o755)
	return err
}

func (ofs *osOutputFS) Rename(oldpath string, newpath string) error {
	if !strings.HasPrefix(newpath, ofs.baseDir) {
		newpath = path.Join(ofs.baseDir, newpath)
	}

	return os.Rename(oldpath, newpath)
}

func (ofs *osOutputFS) Remove(p string) error {
	return os.Remove(path.Join(ofs.baseDir, p))
}

func (ofs *osOutputFS) RemoveAll(p string) error {
	return os.RemoveAll(path.Join(ofs.baseDir, p))
}

func (ofs *osOutputFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return ofs.ReadFileFS.(fs.ReadDirFS).ReadDir(name)
}

func (ofs *osOutputFS) MkdirTemp(pattern string) (OutputFS, string, error) {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return nil, "", err
	}

	return NewOSOutputFS(dir), dir, nil
}
