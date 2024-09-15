package drydock

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"syscall"
	"testing/fstest"
	"time"
)

// MapFSOutputFS extends [testing/fstest.MapFS] with [OutputFS] capabilities.
type MapFSOutputFS struct {
	fstest.MapFS
	baseDir string
}

var _ OutputFS = (*MapFSOutputFS)(nil)

func (fsys *MapFSOutputFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	name = path.Join(fsys.baseDir, name)
	file := &fstest.MapFile{Mode: perm}
	fsys.MapFS[name] = file

	return &writableMapFSFile{MapFile: file, name: name}, nil
}

func (fsys *MapFSOutputFS) Mkdir(name string) error {
	name = path.Join(fsys.baseDir, name)

	if _, exists := fsys.MapFS[name]; exists {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrExist}
	}

	fsys.MapFS[name] = &fstest.MapFile{Mode: 0o755 | os.ModeDir}

	return nil
}

func (fsys *MapFSOutputFS) Rename(oldpath string, newpath string) error {
	if oldpath == newpath {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: syscall.EEXIST}
	}

	file, ok := fsys.MapFS[oldpath]
	if !ok {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: fs.ErrNotExist}
	}

	delete(fsys.MapFS, oldpath)
	fsys.MapFS[newpath] = file

	return nil
}

func (fsys *MapFSOutputFS) Remove(name string) error {
	name = path.Join(fsys.baseDir, name)

	file, exists := fsys.MapFS[name]
	if !exists {
		return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrNotExist}
	}

	if file.Mode&os.ModeDir != 0 {
		panic("AAAHH")
	}

	delete(fsys.MapFS, name)

	return nil
}

func (fsys *MapFSOutputFS) RemoveAll(name string) error {
	if !strings.HasPrefix(name, "/") {
		name = path.Join(fsys.baseDir, name)
	}

	if _, exists := fsys.MapFS[name]; !exists && name != "." {
		return nil
	}

	toRemove := []string{}
	for p := range fsys.MapFS {
		if p == name || path.Dir(p) == name {
			toRemove = append(toRemove, p)
		}
	}

	for _, r := range toRemove {
		delete(fsys.MapFS, r)
	}

	return nil
}

var mapfsTmpDirCounter atomic.Uint32

func (fsys *MapFSOutputFS) MkdirTemp(pattern string) (OutputFS, string, error) {
	patternIndex := strings.Index(pattern, "*")
	if patternIndex != -1 {
		pattern = pattern[0:patternIndex] + fmt.Sprint(mapfsTmpDirCounter.Add(1)) + pattern[patternIndex+1:]
	}

	dirpath := path.Join("/tmp", pattern)

	fsys.MapFS[dirpath] = &fstest.MapFile{Mode: 0o755 | os.ModeDir}

	return &MapFSOutputFS{MapFS: fsys.MapFS, baseDir: dirpath}, dirpath, nil
}

type writableMapFSFile struct {
	*fstest.MapFile
	name string
	b    bytes.Buffer
}

func (f *writableMapFSFile) Name() string {
	return f.name
}

func (f *writableMapFSFile) Write(p []byte) (int, error) {
	return f.b.Write(p)
}

func (f *writableMapFSFile) Read(p []byte) (int, error) {
	return f.b.Read(p)
}

func (f *writableMapFSFile) Close() error {
	f.Data = f.b.Bytes()
	return nil
}

func (f *writableMapFSFile) Stat() (fs.FileInfo, error) {
	return &mapFileStat{
		name:    f.name,
		size:    int64(len(f.Data)),
		mode:    f.Mode,
		modTime: f.ModTime,
		sys:     f.Sys,
	}, nil
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
