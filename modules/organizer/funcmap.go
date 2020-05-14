package organizer

import (
	"bytes"
	"path/filepath"
	"text/template"
)

func funcmap(t *template.Template) template.FuncMap {
	return template.FuncMap{
		"ext": filepath.Ext,
		//XXX: it is dangerous as it modify the input data structure
		"extend": extendMap,

		"tmpl":     tmplName(t),
		"tmplExec": tmplExec,
		"tmplFile": tmplFile,

		"sanitize": pathSanitizer,
		"nospace":  nospaceSanitizer,
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

func extendMap(m map[string]interface{}, key, val string) string {
	m[key] = val
	return ""
}
