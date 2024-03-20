package drydock

import (
	"errors"
	"io"
	"os"
	"text/template"
)

func PlainFile(name string, contents string) File {
	return &plainFile{name: name, contents: []byte(contents)}
}

type plainFile struct {
	name     string
	contents []byte
}

func (f *plainFile) Name() string {
	return f.name
}

// WriteTo implements [io.WriterTo]
func (f *plainFile) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(f.contents)
	if err != nil {
		return 0, err
	}

	return int64(n), nil
}

func TemplateFile(name string, template string, data any) File {
	return &tmplFile{name: name, template: template, data: data}
}

type tmplFile struct {
	name     string
	template string
	data     any
}

func (f *tmplFile) Name() string {
	return f.name
}

// WriteTo implements [io.WriterTo]
func (f *tmplFile) WriteTo(w io.Writer) (int64, error) {
	t, err := template.New(f.name).Parse(f.template)
	if err != nil {
		return 0, err
	}

	return 0, t.Execute(w, f.data)
}

type Template interface {
	Execute(w io.Writer, data any) error
}

func FileFromTemplate(name string, template Template, data any) File {
	return &fileFromTmpl{name: name, template: template, data: data}
}

type fileFromTmpl struct {
	name     string
	template Template
	data     any
}

func (f *fileFromTmpl) Name() string {
	return f.name
}

// WriteTo implements [io.WriterTo]
func (f *fileFromTmpl) WriteTo(w io.Writer) (int64, error) {
	return 0, f.template.Execute(w, f.data)
}

func ModifyFile[E any](name string, parse func([]byte, any) error, mod func(*E) ([]byte, error)) File {
	return &modFile[E]{
		name:  name,
		parse: parse,
		mod:   mod,
	}
}

type modFile[E any] struct {
	name  string
	parse func([]byte, any) error
	mod   func(*E) ([]byte, error)
}

func (f *modFile[E]) Name() string {
	return f.name
}

func (f *modFile[E]) IsNewFile() bool {
	return false
}

// WriteTo implements [WriterToFile]
func (f *modFile[E]) WriteToFile(rootFS WritableFS, filename string, w io.Writer) (int64, error) {
	contents, err := rootFS.ReadFile(filename)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return 0, err
		}
	}

	var e E
	if len(contents) != 0 {
		err = f.parse(contents, &e)
		if err != nil {
			return 0, err
		}
	}

	modified, err := f.mod(&e)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(modified)
	if err != nil {
		return 0, err
	}

	return int64(n), nil
}
