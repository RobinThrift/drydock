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
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestFSGenerator_Generate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := WritableMapFS{}

	g := &FSGenerator{FS: tmpdir}

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

	rootDir := readDir(tmpdir, ".")
	assert.Len(t, rootDir, 3)

	binDir := readDir(tmpdir, "bin")
	assert.Len(t, binDir, 1)

	pkgDir := readDir(tmpdir, "pkg")
	assert.Len(t, pkgDir, 2)

	pkgCliDir := readDir(tmpdir, "pkg", "cli")
	assert.Len(t, pkgCliDir, 2)

	readme, err := tmpdir.ReadFile("README.md")
	assert.NoError(t, err)
	assert.Equal(t, "This is the package", string(readme))

	binCliMainGo, err := tmpdir.ReadFile(path.Join("bin", "cli", "main.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package main", string(binCliMainGo))

	pkgReadme, err := tmpdir.ReadFile(path.Join("pkg", "README.md"))
	assert.NoError(t, err)
	assert.Equal(t, "how to use this thing", string(pkgReadme))

	pkgCliCliGo, err := tmpdir.ReadFile(path.Join("pkg", "cli", "cli.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package cli...", string(pkgCliCliGo))

	pkgCliRunGo, err := tmpdir.ReadFile(path.Join("pkg", "cli", "run.go"))
	assert.NoError(t, err)
	assert.Equal(t, "package cli...run...", string(pkgCliRunGo))
}

func ExampleFSGenerator_Generate() {
	outpath := path.Join(os.TempDir(), "out")
	outfs, err := MkWritableDirFS(outpath)
	if err != nil {
		panic(err)
	}

	g := &FSGenerator{
		FS: outfs,
	}

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

func TestFSGenerator_Generate_ErrorOnExistingDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := WritableMapFS{}

	g := &FSGenerator{FS: tmpdir}

	err := g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	g = &FSGenerator{FS: tmpdir, ErrorOnExistingDir: true}

	err = g.Generate(ctx, Dir("will_exist"))
	assert.Error(t, err)

	err = g.Generate(ctx, Dir("created_twice"), Dir("created_twice"))
	assert.NoError(t, err)
}

func TestFSGenerator_Generate_ErrorOnExistingFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := WritableMapFS{}

	g := &FSGenerator{FS: tmpdir, ErrorOnExistingFile: true}

	err := g.Generate(ctx, Dir("will_exist", PlainFile("test", "contents")))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("will_exist", PlainFile("test", "contents")))
	assert.ErrorIs(t, err, fs.ErrExist)
}

func TestFSGenerator_Generate_CleanDir(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := WritableMapFS{}

	g := &FSGenerator{
		FS:                 tmpdir,
		CleanDir:           true,
		ErrorOnExistingDir: true,
	}

	err := g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("will_exist"))
	assert.NoError(t, err)

	err = g.Generate(ctx, Dir("different_dir"))
	assert.NoError(t, err)

	entries := readDir(tmpdir, ".")

	assert.Len(t, entries, 1)
	assert.Equal(t, "different_dir", entries[0].Name())
	assert.True(t, entries[0].IsDir())
}

func TestFSGenerator_Generate_FileFromTemplate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := WritableMapFS{}

	g := &FSGenerator{ //nolint:varnamelen // This is just a test
		FS:                 tmpdir,
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

	testJsonContents, err := tmpdir.ReadFile("test.json")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(testJsonContents))

	data2 := struct{ Foo string }{"Bar"}

	err = g.Generate(
		ctx,
		FileFromTemplate("test2.json", tmplt, data2),
	)
	assert.Error(t, err)
}

func TestFSGenerator_Generate_ModifyFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tmpdir := WritableMapFS{}

	g := &FSGenerator{FS: tmpdir, CleanDir: false, ErrorOnExistingFile: true} //nolint: varnamelen

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
			c.Baz = "added"
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

	configJSON, err := tmpdir.ReadFile("config.json")
	assert.NoError(t, err)
	assert.Equal(t, `{"foo":"bar","baz":"added"}`, string(configJSON))

	configINI, err := tmpdir.ReadFile(path.Join(".config", "config.ini"))
	assert.NoError(t, err)
	assert.Equal(t, "foo = modified\nbaz = added\n", string(configINI))
}

func readDir(wmfs WritableMapFS, p ...string) []fs.FileInfo {
	dir := path.Join(p...)
	entries := []fs.FileInfo{}

	for k := range wmfs {
		if path.Dir(k) == dir {
			stat, _ := wmfs.Stat(k)
			entries = append(entries, stat)
		}
	}

	return entries
}
