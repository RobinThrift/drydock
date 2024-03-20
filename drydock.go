package drydock

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
)

type File interface {
	Name() string
}

type Directory interface {
	File
	Entries() ([]File, error)
}

type WriterToFile interface {
	WriteToFile(rootFS WritableFS, filename string, w io.Writer) (n int64, err error)
}

type IsNewFile interface {
	IsNewFile() bool
}

type writerToAdapter struct {
	io.WriterTo
}

func (a *writerToAdapter) WriteToFile(_ WritableFS, _ string, w io.Writer) (n int64, err error) {
	return a.WriteTo(w)
}

var ErrCleaningOutputDir = errors.New("error cleaning output dir")

func cleanDir(rootFS WritableFS, dir string) error {
	f, err := rootFS.Open(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return nil
	}

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCleaningOutputDir, err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%w: can't clean dir %s: %[1]s is not a directory", ErrCleaningOutputDir, dir)
	}

	if dirFS, ok := rootFS.(fs.ReadDirFS); ok {
		entries, readDirErr := dirFS.ReadDir(dir)
		if readDirErr != nil {
			return fmt.Errorf("%w: %w", ErrCleaningOutputDir, readDirErr)
		}

		for _, e := range entries {
			err = rootFS.RemoveAll(path.Join(dir, e.Name()))
			if err != nil {
				return fmt.Errorf("%w: %w", ErrCleaningOutputDir, err)
			}
		}

		return nil
	}

	if dirFS, ok := rootFS.(fs.ReadDirFile); ok {
		entries, readDirErr := dirFS.ReadDir(0)
		if readDirErr != nil {
			return fmt.Errorf("%w: %w", ErrCleaningOutputDir, readDirErr)
		}

		for _, e := range entries {
			err = rootFS.RemoveAll(path.Join(dir, e.Name()))
			if err != nil {
				return fmt.Errorf("%w: %w", ErrCleaningOutputDir, err)
			}
		}

		return nil
	}

	err = rootFS.RemoveAll(".")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCleaningOutputDir, err)
	}

	return nil
}
