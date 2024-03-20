package drydock

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"syscall"
	"testing/fstest"
	"time"
)

// WritableMapFS extends [testing/fstest.MapFS] with [WritableFS] capabilities.
type WritableMapFS fstest.MapFS

var _ WritableFS = (*WritableMapFS)(nil)

func (fsys WritableMapFS) Glob(pattern string) ([]string, error) {
	return fstest.MapFS(fsys).Glob(pattern)
}

func (fsys WritableMapFS) Open(name string) (fs.File, error) {
	return fstest.MapFS(fsys).Open(name)
}

func (fsys WritableMapFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fstest.MapFS(fsys).ReadDir(name)
}

func (fsys WritableMapFS) ReadFile(name string) ([]byte, error) {
	return fstest.MapFS(fsys).ReadFile(name)
}

func (fsys WritableMapFS) Stat(name string) (fs.FileInfo, error) {
	return fstest.MapFS(fsys).Stat(name)
}

func (fsys WritableMapFS) Sub(dir string) (fs.FS, error) {
	return fstest.MapFS(fsys).Sub(dir)
}

func (fsys WritableMapFS) Mkdir(name string, perm fs.FileMode) error {
	if _, exists := fsys[name]; exists {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrExist}
	}

	fsys[name] = &fstest.MapFile{Mode: perm | os.ModeDir}

	return nil
}

func (fsys WritableMapFS) Rename(oldpath string, newpath string) error {
	if oldpath == newpath {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: syscall.EEXIST}
	}

	file, ok := fsys[oldpath]
	if !ok {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath}
	}

	delete(fsys, oldpath)
	fsys[newpath] = file

	return nil
}

func (fsys WritableMapFS) Remove(path string) error {
	if _, exists := fsys[path]; exists {
		return &fs.PathError{Op: "remove", Path: path, Err: fs.ErrExist}
	}

	delete(fsys, path)

	return nil
}

func (fsys WritableMapFS) RemoveAll(dir string) error {
	if _, exists := fsys[dir]; !exists && dir != "." {
		return &fs.PathError{Op: "remove", Path: dir, Err: fs.ErrExist}
	}

	toRemove := []string{}
	for p := range fsys {
		if p == dir || path.Dir(p) == dir {
			toRemove = append(toRemove, p)
		}
	}

	for _, r := range toRemove {
		delete(fsys, r)
	}

	return nil
}

// CreateTemp does not implement the pattern function of [os.CreateTemp].
// The default temp dir is `/tmp`.
func (fsys WritableMapFS) CreateTemp(dir string, pattern string) (WritableFile, error) {
	if dir == "" {
		dir = "/tmp"
	}

	name := path.Join(dir, pattern)

	file := &fstest.MapFile{Mode: 0644}

	fsys[name] = file

	return &writableMapFile{name: name, f: file}, nil
}

type writableMapFile struct {
	name string
	f    *fstest.MapFile
	b    bytes.Buffer
}

func (wf *writableMapFile) Name() string {
	return wf.name
}

func (wf *writableMapFile) Write(p []byte) (int, error) {
	return wf.b.Write(p)
}

func (wf *writableMapFile) Stat() (fs.FileInfo, error) {
	return &mapFileStat{
		name:    wf.name,
		size:    int64(len(wf.f.Data)),
		mode:    wf.f.Mode,
		modTime: wf.f.ModTime,
		sys:     wf.f.Sys,
	}, nil
}

func (wf *writableMapFile) Read(p []byte) (int, error) {
	return wf.b.Read(p)
}

func (wf *writableMapFile) Close() error {
	wf.f.Data = wf.b.Bytes()
	return nil
}

type mapFileStat struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	sys     any
}

func (fs *mapFileStat) Name() string       { return fs.name }
func (fs *mapFileStat) Size() int64        { return fs.size }
func (fs *mapFileStat) IsDir() bool        { return fs.mode.IsDir() }
func (fs *mapFileStat) Mode() fs.FileMode  { return fs.mode }
func (fs *mapFileStat) ModTime() time.Time { return fs.modTime }
func (fs *mapFileStat) Sys() any           { return fs.sys }
