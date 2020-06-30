package util

import (
	"bytes"
	"text/template"
)

// FuncMap exports a set of functions in template.FuncMap format:
// - tmpl:     execute a sub-template by name. Sub-templates are chosen from t namespace
// - tmplExec: execute a template text
// - tmplFile: execute a template stored in a file
func FuncMap(t *template.Template) template.FuncMap {
	return template.FuncMap{
		"tmpl":     tmplName(t),
		"tmplExec": tmplExec,
		"tmplFile": tmplFile,
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
