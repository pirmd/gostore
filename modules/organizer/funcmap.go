package organizer

import (
	"bytes"
	"path/filepath"
	"text/template"
)

func (o *organizer) funcmap() template.FuncMap {
	return template.FuncMap{
		"ext": filepath.Ext,

		"tmpl":     tmplName(o.namers),
		"tmplExec": tmplExec,
		"tmplFile": tmplFile,

		"sanitizePath": pathSanitizer,
		"nospace":      nospaceSanitizer,
	}
}

func tmplName(t *template.Template) func(string, interface{}) (string, error) {
	return func(name string, v interface{}) (string, error) {
		buf := &bytes.Buffer{}
		err := t.ExecuteTemplate(buf, name, v)
		return buf.String(), err
	}
}

func tmplExec(src string, v interface{}) (string, error) {
	t, err := template.New("temp").Parse(src)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, v)
	return buf.String(), err
}

func tmplFile(name string, v interface{}) (string, error) {
	t, err := template.ParseFiles(name)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, v)
	return buf.String(), err
}
