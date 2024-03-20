package drydock

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestDirFSGenerator_Generate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{OutputDir: tmpdir}

	err := g.Generate(ctx,
		PlainFile("README.md", "This is the package"),
		Dir("bin",
			Dir("cli",
				PlainFile("main.go", "package main"),
			),
		),
		Dir("pkg",
			PlainFile("README.md", "how to use this thing"),
			Dir("cli",
				PlainFile("cli.go", "package cli..."),
				PlainFile("run.go", "package cli...run..."),
			),
		),
	)

	assert.NoError(t, err)

	rootDir, err := os.ReadDir(tmpdir)
	assert.NoError(t, err)
	assert.Len(t, rootDir, 3)

	binDir, err := os.ReadDir(path.Join(tmpdir, "bin"))
	assert.NoError(t, err)
	assert.Len(t, binDir, 1)

	pkgDir, err := os.ReadDir(path.Join(tmpdir, "pkg"))
	assert.NoError(t, err)
	assert.Len(t, pkgDir, 2)

	pkgCliDir, err := os.ReadDir(path.Join(tmpdir, "pkg", "cli"))
	assert.NoError(t, err)
	assert.Len(t, pkgCliDir, 2)

	readme, err := os.ReadFile(path.Join(tmpdir, "README.md"))
	assert.NoError(t, err)
	assert.Equal(t, "This is the package", string(readme))

	binCliMainGo, err := os.ReadFile(path.Join(tmpdir, "bin", "cli", "main.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package main", string(binCliMainGo))

	pkgReadme, err := os.ReadFile(path.Join(tmpdir, "pkg", "README.md"))
	assert.NoError(t, err)
	assert.Equal(t, "how to use this thing", string(pkgReadme))

	pkgCliCliGo, err := os.ReadFile(path.Join(tmpdir, "pkg", "cli", "cli.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package cli...", string(pkgCliCliGo))

	pkgCliRunGo, err := os.ReadFile(path.Join(tmpdir, "pkg", "cli", "run.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package cli...run...", string(pkgCliRunGo))
}

func TestDirFSGenerator_Generate_ErrorOnExistingDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{OutputDir: tmpdir}

	err := g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	g = &DirFSGenerator{OutputDir: tmpdir, ErrorOnExistingDir: true}

	err = g.Generate(ctx, Dir("will_exist"))
	assert.Error(t, err)

	err = g.Generate(ctx, Dir("created_twice"), Dir("created_twice"))
	assert.NoError(t, err)

}

func TestDirFSGenerator_Generate_ErrorOnExistingFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{OutputDir: tmpdir, ErrorOnExistingFile: true}

	err := g.Generate(ctx, Dir("will_exist", PlainFile("test", "contents")))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("will_exist", PlainFile("test", "contents")))
	assert.ErrorIs(t, err, os.ErrExist)
}

func TestDirFSGenerator_Generate_CleanDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{
		OutputDir:          tmpdir,
		CleanDir:           true,
		ErrorOnExistingDir: true,
	}

	err := g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("different_dir"))
	assert.NoError(t, err)

	entries, err := os.ReadDir(tmpdir)
	assert.NoError(t, err)

	assert.Len(t, entries, 1)
	assert.Equal(t, "different_dir", entries[0].Name())
	assert.True(t, entries[0].IsDir())
}

func TestDirFSGenerator_Generate_CleanDir_DirNotExist(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{
		OutputDir: path.Join(tmpdir, "non_existent"),
		CleanDir:  true,
	}

	err := g.Generate(ctx, PlainFile("test.txt", "contents"))
	assert.NoError(t, err)
}

func TestDirFSGenerator_Generate_FileFromTemplate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{ //nolint:varnamelen // This is just a test
		OutputDir:          tmpdir,
		CleanDir:           true,
		ErrorOnExistingDir: true,
	}

	data := map[string]any{
		"foo": "bar",
		"baz": "bat",
	}

	tmplt, err := template.New("").Parse(`{
		"foo": "{{ .foo }}",
		"baz": "{{ .baz }}",
}`)
	assert.NoError(t, err)

	err = g.Generate(
		ctx,
		FileFromTemplate("test.json", tmplt, data),
	)
	assert.NoError(t, err)

	expected := `{
		"foo": "bar",
		"baz": "bat",
}`

	testJsonContents, err := os.ReadFile(path.Join(tmpdir, "test.json"))
	assert.NoError(t, err)
	assert.Equal(t, expected, string(testJsonContents))

	data2 := struct{ Foo string }{"Bar"}

	err = g.Generate(
		ctx,
		FileFromTemplate("test2.json", tmplt, data2),
	)
	assert.Error(t, err)
}

func TestDirFSGenerator_Generate_ModifyFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{OutputDir: tmpdir, CleanDir: false, ErrorOnExistingFile: true} //nolint: varnamelen

	err := g.Generate(
		ctx,
		PlainFile("config.json", `{"foo": "bar"}`),
		Dir(".config",
			PlainFile("config.ini", `foo = bar`),
		),
	)

	assert.NoError(t, err)

	type config struct {
		Foo string `json:"foo"`
		Baz string `json:"baz"`
	}

	unmarshalINI := func(d []byte, target any) error {
		c, ok := target.(*config)
		if !ok {
			panic("parseINI can only decode config")
		}

		lines := bytes.Split(d, []byte("\n"))
		for _, l := range lines {
			pairs := bytes.Split(l, []byte("="))
			key := string(bytes.TrimSpace(pairs[0]))
			value := bytes.TrimSpace(pairs[1])
			if key == "foo" {
				c.Foo = string(value)
				continue
			}

			if key == "baz" {
				c.Baz = string(value)
				continue
			}
		}
		return nil
	}

	marshalINI := func(c *config) ([]byte, error) {
		var b bytes.Buffer

		if c.Foo != "" {
			b.WriteString("foo = ")
			b.WriteString(c.Foo)
			b.WriteString("\n")
		}

		if c.Baz != "" {
			b.WriteString("baz = ")
			b.WriteString(c.Baz)
			b.WriteString("\n")
		}

		return b.Bytes(), nil
	}

	err = g.Generate(
		ctx,
		ModifyFile("config.json", json.Unmarshal, func(c *config) ([]byte, error) {
			c.Baz = "added" //nolint: goconst // for tests
			return json.Marshal(c)
		}),
		Dir(".config",
			ModifyFile("config.ini", unmarshalINI, func(c *config) ([]byte, error) {
				c.Foo = "modified"
				c.Baz = "added"
				return marshalINI(c)
			}),
		),
	)
	assert.NoError(t, err)

	configJSON, err := os.ReadFile(path.Join(tmpdir, "config.json"))
	assert.NoError(t, err)
	assert.Equal(t, `{"foo":"bar","baz":"added"}`, string(configJSON))

	configINI, err := os.ReadFile(path.Join(tmpdir, ".config", "config.ini"))
	assert.NoError(t, err)
	assert.Equal(t, "foo = modified\nbaz = added\n", string(configINI))
}

func TestFSGenerator_Generate_NoCreateOutputDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := t.TempDir()

	g := &DirFSGenerator{
		OutputDir:         path.Join(tmpdir, "will_not_create"),
		NoCreateOutputDir: true,
	}

	err := g.Generate(ctx, PlainFile("test.txt", "contents"))
	assert.Error(t, err)

	err = os.Mkdir(g.OutputDir, 0755)
	assert.NoError(t, err)

	err = g.Generate(ctx, PlainFile("test.txt", "contents"))
	assert.NoError(t, err)

	actual, err := os.ReadFile(path.Join(tmpdir, "will_not_create", "test.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "contents", string(actual))
}
