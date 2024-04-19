package drydock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
)

var ErrMissingFS = errors.New("missing FS")

type FSGenerator struct {
	FS WritableFS

	ErrorOnExistingDir  bool
	CleanDir            bool
	ErrorOnExistingFile bool

	createdDirs map[string]struct{}
}

var createdDir = struct{}{}

type genfile struct {
	path      string
	contents  WriterToFile
	isNewFile bool
}

type gendir struct {
	path string
}

func (g *FSGenerator) Generate(ctx context.Context, files ...File) error {
	if g.FS == nil {
		return ErrMissingFS
	}

	g.createdDirs = map[string]struct{}{}

	if g.CleanDir {
		err := cleanDir(g.FS, ".")
		if err != nil {
			return err
		}
	}

	gendirs := make([]*gendir, 0, len(files))
	genfiles := make([]*genfile, 0, len(files))

	for _, f := range files {
		dirs, files, err := g.generate(ctx, "", f)
		if err != nil {
			return err
		}

		gendirs = append(gendirs, dirs...)
		genfiles = append(genfiles, files...)
	}

	for _, d := range gendirs {
		err := g.generateRealDir(d.path)
		if err != nil {
			return err
		}
	}

	for _, f := range genfiles {
		err := g.generateRealFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *FSGenerator) generateRealDir(dir string) error {
	err := g.FS.Mkdir(dir, 0755)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}

		_, created := g.createdDirs[dir]
		if g.ErrorOnExistingDir && !created {
			return err
		}
	}

	g.createdDirs[dir] = createdDir

	return nil
}

func (g *FSGenerator) generateRealFile(file *genfile) error {
	if g.ErrorOnExistingFile && file.isNewFile {
		stat, statErr := statFile(g.FS, file.path)
		if statErr != nil {
			if !errors.Is(statErr, fs.ErrNotExist) {
				return statErr
			}
		}

		if stat != nil {
			return fmt.Errorf("file already exits %s: %w", file.path, fs.ErrExist)
		}
	}

	tmpfile, err := g.FS.CreateTemp("", path.Base(file.path))
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(
			err,
			tmpfile.Close(),
		)

		if err != nil {
			err = errors.Join(err, g.FS.Remove(tmpfile.Name()))
		}
	}()

	_, err = file.contents.WriteToFile(g.FS, file.path, tmpfile)
	if err != nil {
		return err
	}

	err = g.FS.Rename(tmpfile.Name(), file.path)
	if err != nil {
		return fmt.Errorf("error moving tempfile to real file %s: %w", file.path, err)
	}

	return nil
}

func (g *FSGenerator) generate(ctx context.Context, parentDir string, file File) ([]*gendir, []*genfile, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	if dir, ok := file.(Directory); ok {
		return g.generateDir(ctx, parentDir, dir)
	}

	isNewFile := true
	if f, ok := file.(IsNewFile); ok {
		isNewFile = f.IsNewFile()
	}

	if wt, ok := file.(WriterToFile); ok {
		return nil, []*genfile{{path: path.Join(parentDir, file.Name()), contents: wt, isNewFile: isNewFile}}, nil
	}

	if wt, ok := file.(io.WriterTo); ok {
		return nil, []*genfile{{path: path.Join(parentDir, file.Name()), contents: &writerToAdapter{wt}, isNewFile: isNewFile}}, nil
	}

	return nil, nil, nil
}

func (g *FSGenerator) generateDir(ctx context.Context, parentDir string, dir Directory) ([]*gendir, []*genfile, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	dirpath := path.Join(parentDir, dir.Name())

	entries, err := dir.Entries()
	if err != nil {
		return nil, nil, err
	}

	gendirs := []*gendir{{path: dirpath}}
	genfiles := make([]*genfile, 0, len(entries))

	for _, f := range entries {
		dirs, files, err := g.generate(ctx, dirpath, f)
		if err != nil {
			return nil, nil, err
		}

		gendirs = append(gendirs, dirs...)
		genfiles = append(genfiles, files...)
	}

	return gendirs, genfiles, nil
}

func statFile(rootFS fs.FS, name string) (fs.FileInfo, error) {
	if statFS, ok := rootFS.(fs.StatFS); ok {
		return statFS.Stat(name)
	}

	f, err := rootFS.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Stat()
}
