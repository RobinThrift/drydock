package drydock

import (
	"io"
	"text/template"
)

// A PlainFile with the given contents.
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

type Template interface {
	Execute(w io.Writer, data any) error
}

// TemplatedFile will execute the template with the given data and write the resuls to the output file.
func TemplatedFile(name string, template Template, data any) File {
	return &templatedFile{name: name, template: template, data: data}
}

type tmplfunc func(w io.Writer, data any) error

func (f tmplfunc) Execute(w io.Writer, data any) error {
	return f(w, data)
}

func TemplatedFileStr(name string, templateStr string, data any) File {
	return &templatedFile{name: name, template: tmplfunc(func(w io.Writer, data any) error {
		t, err := template.New(name).Parse(templateStr)
		if err != nil {
			return err
		}

		return t.Execute(w, data)
	}), data: data}
}

type templatedFile struct {
	name     string
	template Template
	data     any
}

func (f *templatedFile) Name() string {
	return f.name
}

// WriteTo implements [io.WriterTo]
func (f *templatedFile) WriteTo(w io.Writer) (int64, error) {
	return 0, f.template.Execute(w, f.data)
}

// ModifyFile can modify an existing file's contents.
// [Generator.Generate] will return an error, if the file doesn't exist yet.
func ModifyFile(name string, modifier func(contents []byte, w io.Writer) error) File {
	return &modFile{
		name:     name,
		modifier: modifier,
	}
}

type modFile struct {
	name     string
	modifier func(contents []byte, w io.Writer) error
}

func (f *modFile) Name() string {
	return f.name
}

// WriteModifiedTo implements [WriterToModify].
func (f *modFile) WriteModifiedTo(contents []byte, w io.Writer) error {
	return f.modifier(contents, w)
}

func ModifyMarshalledFunc[V any](unmarshal func([]byte, any) error, modify func(*V) ([]byte, error)) func(contents []byte, w io.Writer) error {
	return func(contents []byte, w io.Writer) error {
		var val V
		err := unmarshal(contents, &val)
		if err != nil {
			return err
		}

		modified, err := modify(&val)
		if err != nil {
			return err
		}

		_, err = w.Write(modified)
		return err
	}
}
