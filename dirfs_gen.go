package drydock

import (
	"context"
)

type DirFSGenerator struct {
	OutputDir           string
	ErrorOnExistingDir  bool
	NoCreateOutputDir   bool
	CleanDir            bool
	ErrorOnExistingFile bool
}

func (g *DirFSGenerator) Generate(ctx context.Context, files ...File) error {
	var fs WritableFS
	var err error

	if g.NoCreateOutputDir {
		fs = NewWritableDirFS(g.OutputDir)
	} else {
		fs, err = MkWritableDirFS(g.OutputDir)
	}

	if err != nil {
		return err
	}

	fsgen := &FSGenerator{
		FS:                  fs,
		CleanDir:            g.CleanDir,
		ErrorOnExistingDir:  g.ErrorOnExistingDir,
		ErrorOnExistingFile: g.ErrorOnExistingFile,
	}

	return fsgen.Generate(ctx, files...)
}
