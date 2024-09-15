package drydock

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
)

var ErrCleaningOutputDir = errors.New("error cleaning output dir")

func cleanDir(rootFS OutputFS, dir string) error {
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

	var entries []fs.DirEntry

	switch dirFS := rootFS.(type) {
	case fs.ReadDirFS:
		entries, err = dirFS.ReadDir(dir)
	case fs.ReadDirFile:
		entries, err = dirFS.ReadDir(0)
	default:
		err = rootFS.RemoveAll(".")
	}

	if err != nil {
		return fmt.Errorf("%w: %w", ErrCleaningOutputDir, err)
	}

	for _, e := range entries {
		if e.Name() == "." {
			continue
		}

		err = rootFS.RemoveAll(path.Join(dir, e.Name()))
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCleaningOutputDir, err)
		}
	}

	return nil
}

func fileExists(rootFS fs.FS, name string) (bool, error) {
	f, err := rootFS.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return true, f.Close()
}
