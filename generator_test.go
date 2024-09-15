package drydock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"
	"testing/fstest"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_Generate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpfs := &MapFSOutputFS{MapFS: fstest.MapFS{}, baseDir: "."}
	g := NewGenerator(tmpfs)

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
	require.NoError(t, err)

	rootDir := readDir(tmpfs, ".")
	assert.Len(t, rootDir, 3)

	binDir := readDir(tmpfs, "bin")
	assert.Len(t, binDir, 1)

	pkgDir := readDir(tmpfs, "pkg")
	assert.Len(t, pkgDir, 2)

	pkgCliDir := readDir(tmpfs, "pkg", "cli")
	assert.Len(t, pkgCliDir, 2)

	readme, err := tmpfs.ReadFile("README.md")
	assert.NoError(t, err)
	assert.Equal(t, "This is the package", string(readme))

	binCliMainGo, err := tmpfs.ReadFile(path.Join("bin", "cli", "main.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package main", string(binCliMainGo))

	pkgReadme, err := tmpfs.ReadFile(path.Join("pkg", "README.md"))
	assert.NoError(t, err)
	assert.Equal(t, "how to use this thing", string(pkgReadme))

	pkgCliCliGo, err := tmpfs.ReadFile(path.Join("pkg", "cli", "cli.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package cli...", string(pkgCliCliGo))

	pkgCliRunGo, err := tmpfs.ReadFile(path.Join("pkg", "cli", "run.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package cli...run...", string(pkgCliRunGo))
}

func TestGenerator_Generate_ErrorOnExistingDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpfs := &MapFSOutputFS{MapFS: fstest.MapFS{}, baseDir: "."}
	g := NewGenerator(tmpfs, WithErrorOnExistingDir(false))

	err := g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	g = NewGenerator(tmpfs, WithErrorOnExistingDir(true))

	err = g.Generate(ctx, Dir("will_exist"))
	assert.Error(t, err)

	g = NewGenerator(tmpfs, WithErrorOnExistingDir(true))

	err = g.Generate(ctx, Dir("created_twice"), Dir("created_twice"))
	assert.NoError(t, err)
}

func TestGenerator_Generate_ErrorOnExistingFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpfs := &MapFSOutputFS{MapFS: fstest.MapFS{}, baseDir: "."}
	g := NewGenerator(tmpfs, WithErrorOnExistingFile(false))

	err := g.Generate(ctx, Dir("will_exist", PlainFile("test", "contents")))
	assert.NoError(t, err)

	t.Run("No Error", func(t *testing.T) {
		g := NewGenerator(tmpfs, WithErrorOnExistingFile(false))
		err := g.Generate(ctx, Dir("will_exist", PlainFile("test", "contents")))
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		g := NewGenerator(tmpfs, WithErrorOnExistingFile(true))
		err := g.Generate(ctx, Dir("will_exist", PlainFile("test", "contents")))
		assert.ErrorIs(t, err, fs.ErrExist)
	})
}

func TestGenerator_Generate_EmptyOutputDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpfs := &MapFSOutputFS{MapFS: fstest.MapFS{}, baseDir: "."}
	g := NewGenerator(tmpfs, WithErrorOnExistingDir(true), WithEmptyOutputDir(true))

	err := g.Generate(ctx, Dir("will_exist"))
	require.NoError(t, err)

	g = NewGenerator(tmpfs, WithErrorOnExistingDir(true), WithEmptyOutputDir(true))

	err = g.Generate(ctx, Dir("will_exist"))
	require.NoError(t, err)

	g = NewGenerator(tmpfs, WithErrorOnExistingDir(true), WithEmptyOutputDir(true))

	err = g.Generate(ctx, Dir("different_dir"))
	require.NoError(t, err)

	entries := readDir(tmpfs, ".")

	assert.Len(t, entries, 1)
	assert.Equal(t, "different_dir", entries[0].Name())
	assert.True(t, entries[0].IsDir())
}

func TestGenerator_Generate_FileFromTemplate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpfs := &MapFSOutputFS{MapFS: fstest.MapFS{}, baseDir: "."}
	g := NewGenerator(tmpfs, WithErrorOnExistingDir(true), WithEmptyOutputDir(true)) //nolint:varnamelen // This is just a test

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
		TemplatedFile("test.json", tmplt, data),
	)
	assert.NoError(t, err)

	expected := `{
			"foo": "bar",
			"baz": "bat",
	}`

	testJsonContents, err := tmpfs.ReadFile("test.json")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(testJsonContents))

	data2 := struct{ Foo string }{"Bar"}

	err = g.Generate(
		ctx,
		TemplatedFile("test2.json", tmplt, data2),
	)
	assert.Error(t, err)
}

func TestGenerator_Generate_ModifyFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpfs := &MapFSOutputFS{MapFS: fstest.MapFS{}, baseDir: "."}
	g := NewGenerator(tmpfs, WithEmptyOutputDir(true)) //nolint:varnamelen // This is just a test

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

	g = NewGenerator(tmpfs, WithEmptyOutputDir(true))
	err = g.Generate(
		ctx,
		ModifyFile("config.json", ModifyMarshalledFunc(json.Unmarshal, func(c *config) ([]byte, error) {
			c.Baz = "added"
			return json.Marshal(c)
		})),
		Dir(".config",
			ModifyFile("config.ini", ModifyMarshalledFunc(unmarshalINI, func(c *config) ([]byte, error) {
				c.Foo = "modified"
				c.Baz = "added"
				return marshalINI(c)
			})),
		),
	)
	assert.NoError(t, err)

	configJSON, err := tmpfs.ReadFile("config.json")
	assert.NoError(t, err)
	assert.Equal(t, `{"foo":"bar","baz":"added"}`, string(configJSON))

	configINI, err := tmpfs.ReadFile(path.Join(".config", "config.ini"))
	assert.NoError(t, err)
	assert.Equal(t, "foo = modified\nbaz = added\n", string(configINI))
}

func readDir(wmfs *MapFSOutputFS, p ...string) []fs.FileInfo {
	dir := path.Join(p...)
	entries := []fs.FileInfo{}

	for k := range wmfs.MapFS {
		if path.Dir(k) == dir {
			stat, _ := wmfs.Stat(k)
			entries = append(entries, stat)
		}
	}

	return entries
}

func ExampleGenerator_Generate() {
	outpath, err := os.MkdirTemp("", "drydock.ExampleGenerator_Generate-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(outpath)

	outfs := NewOSOutputFS(outpath)

	g := NewGenerator(outfs)

	err = g.Generate(
		context.Background(),
		PlainFile("README.md", "# drydock"),
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

	if err != nil {
		panic(err)
	}

	entries, err := os.ReadDir(outpath)
	if err != nil {
		panic(err)
	}

	for _, e := range entries {
		fmt.Println(e)
	}

	// Output:
	// - README.md
	// d bin/
	// d pkg/
}
