package drydock

import (
	"fmt"
	"io"
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

func (ofs *osOutputFS) Move(oldpath string, newpath string) error {
	if !strings.HasPrefix(newpath, ofs.baseDir) {
		newpath = path.Join(ofs.baseDir, newpath)
	}

	err := os.Rename(oldpath, newpath)
	if err != nil {
		if strings.Contains(err.Error(), "cross-device link") {
			return ofs.copyMove(oldpath, newpath)
		}

		return err
	}

	return nil
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

func (ofs *osOutputFS) copyMove(oldpath string, newpath string) error {
	stat, err := os.Stat(oldpath)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return ofs.copyMoveDir(oldpath, newpath)
	}

	return ofs.copyMoveFile(oldpath, newpath)
}

func (ofs *osOutputFS) copyMoveFile(oldpath string, newpath string) error {
	src, err := os.Open(oldpath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", oldpath, err)
	}
	defer src.Close()

	tmpFilePath := newpath + ".tmp"

	dst, err := os.Create(tmpFilePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", tmpFilePath, err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		_ = dst.Close()
		_ = os.Remove(tmpFilePath)

		return fmt.Errorf("error copying file from %s to %s: %w", oldpath, tmpFilePath, err)
	}

	_ = dst.Close()

	return os.Rename(tmpFilePath, newpath)
}

func (ofs *osOutputFS) copyMoveDir(oldpath string, newpath string) error {
	tmpDirPath := newpath + ".tmp"

	err := os.MkdirAll(tmpDirPath, 0o755)
	if err != nil {
		return fmt.Errorf("error creating dir %s: %w", tmpDirPath, err)
	}

	entries, err := os.ReadDir(oldpath)
	if err != nil {
		return fmt.Errorf("error reading dir %s: %w", oldpath, err)
	}

	for _, entry := range entries {
		entryOldPath := path.Join(oldpath, entry.Name())
		entryNewPath := path.Join(tmpDirPath, entry.Name())
		if entry.IsDir() {
			err = ofs.copyMoveDir(entryOldPath, entryNewPath)
			if err != nil {
				return err
			}
			continue
		}

		err = ofs.copyMoveFile(entryOldPath, entryNewPath)
		if err != nil {
			return err
		}
	}

	return os.Rename(tmpDirPath, newpath)
}
