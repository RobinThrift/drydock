package drydock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

type Generator struct {
	errorOnExistingDir  bool
	emptyOutputDir      bool
	errorOnExistingFile bool

	output OutputFS
	files  []File

	tmptfs      OutputFS
	tmpdir      string
	tmpdirs     map[string]int
	tmpfiles    []string
	tmpmodified []string
}

type Option func(g *Generator)

func WithErrorOnExistingDir(b bool) Option {
	return func(g *Generator) {
		g.errorOnExistingDir = b
	}
}

func WithErrorOnExistingFile(b bool) Option {
	return func(g *Generator) {
		g.errorOnExistingFile = b
	}
}

func WithEmptyOutputDir(b bool) Option {
	return func(g *Generator) {
		g.emptyOutputDir = b
	}
}

func NewGenerator(output OutputFS, opts ...Option) *Generator {
	g := &Generator{
		errorOnExistingDir:  false,
		errorOnExistingFile: true,
		emptyOutputDir:      false,
		output:              output,
		tmpdirs:             make(map[string]int),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func (g *Generator) Add(files ...File) *Generator {
	g.files = append(g.files, files...)
	return g
}

func (g *Generator) Generate(ctx context.Context, files ...File) error {
	clear(g.tmpdirs)
	clear(g.tmpdirs)
	g.tmpfiles = []string{}
	g.tmpmodified = []string{}

	g.files = append(g.files, files...)

	var err error
	g.tmptfs, g.tmpdir, err = g.output.MkdirTemp("drydock-*")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}

	for _, f := range g.files {
		err := g.generate(ctx, "", f)
		if err != nil {
			return errors.Join(err, g.output.RemoveAll(g.tmpdir))
		}
	}

	return g.moveToOutput()
}

func (g *Generator) generate(ctx context.Context, parentDir string, file File) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if file.Name() == "" {
		return fmt.Errorf("invalid filename/dirname")
	}

	if dir, ok := file.(Directory); ok {
		return g.generateDir(ctx, parentDir, dir)
	}

	return g.generateFile(parentDir, file)
}

func (g *Generator) generateDir(ctx context.Context, parentDir string, dir Directory) error {
	dirpath := path.Join(parentDir, dir.Name())

	err := g.tmptfs.Mkdir(dirpath)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("error creating temporary dir '%s': %w", dirpath, err)
		}
	}

	if _, exists := g.tmpdirs[dirpath]; !exists {
		g.tmpdirs[dirpath] = len(g.tmpdirs)
	}

	entries, err := dir.Entries()
	if err != nil {
		return err
	}

	for _, f := range entries {
		err := g.generate(ctx, dirpath, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateFile(parentDir string, file File) error {
	filepath := path.Join(parentDir, file.Name())

	outfile, err := g.tmptfs.OpenFile(filepath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) || g.errorOnExistingFile {
			return fmt.Errorf("error creating temporary file '%s': %w", filepath, err)
		}
	}
	defer outfile.Close()

	outfileWriter, ok := outfile.(io.Writer)
	if !ok {
		return fmt.Errorf("file %s opened with FS %T is not io.Writer", filepath, g.output)
	}

	if modifier, ok := file.(WriterToModify); ok {
		return g.modifyFile(parentDir, file, modifier, outfileWriter)
	}

	wt, ok := file.(io.WriterTo)
	if !ok {
		return nil
	}

	_, err = wt.WriteTo(outfileWriter)
	if err != nil {
		return fmt.Errorf("error writing to temprorar file %s: %w", filepath, err)
	}

	g.tmpfiles = append(g.tmpfiles, filepath)

	return nil
}

func (g *Generator) modifyFile(parentDir string, file File, modifier WriterToModify, outfile io.Writer) error {
	filepath := path.Join(parentDir, file.Name())
	var contents []byte
	var err error

	if readFileFS, ok := g.output.(fs.ReadFileFS); ok {
		contents, err = readFileFS.ReadFile(filepath)
	} else {
		f, openErr := g.output.Open(filepath)
		if openErr != nil {
			return fmt.Errorf("error reading file '%s' for modification: %w", filepath, openErr)
		}
		defer f.Close()
		contents, err = io.ReadAll(f)
	}

	if err != nil {
		return fmt.Errorf("error reading file '%s' for modification: %w", filepath, err)
	}

	err = modifier.WriteModifiedTo(contents, outfile)
	if err != nil {
		return fmt.Errorf("error modifying file '%s': %w", filepath, err)
	}

	g.tmpmodified = append(g.tmpmodified, filepath)

	return nil
}

func (g *Generator) moveToOutput() error {
	defer g.output.RemoveAll(g.tmpdir) //nolint: errcheck

	if g.emptyOutputDir {
		err := cleanDir(g.output, ".")
		if err != nil {
			return err
		}
	}

	tmpdirs := make([]string, len(g.tmpdirs))
	for dir, i := range g.tmpdirs {
		tmpdirs[i] = dir
	}

	for _, dir := range tmpdirs {
		err := g.output.Mkdir(dir)
		if err != nil {
			if !errors.Is(err, fs.ErrExist) || g.errorOnExistingDir {
				return fmt.Errorf("error creating dir '%s': %s", dir, err)
			}
		}
	}

	for _, file := range g.tmpfiles {
		if g.errorOnExistingFile {
			exists, err := fileExists(g.output, file)
			if err != nil {
				return err
			}

			if exists {
				return fmt.Errorf("file already exits %s: %w", file, fs.ErrExist)
			}
		}

		tmpfilepath := path.Join(g.tmpdir, file)

		err := g.output.Rename(tmpfilepath, file)
		if err != nil {
			return fmt.Errorf("error moving file %s to %s: %w", tmpfilepath, file, err)
		}
	}

	for _, file := range g.tmpmodified {
		tmpfilepath := path.Join(g.tmpdir, file)
		err := g.output.Rename(tmpfilepath, file)
		if err != nil {
			return fmt.Errorf("error moving (modified) file %s to %s: %w", tmpfilepath, file, err)
		}
	}

	return nil
}
