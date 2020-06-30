package organizer

import (
	"path/filepath"
	"text/template"

	"github.com/pirmd/gostore/util"
)

func (o *organizer) funcmap() template.FuncMap {
	funcs := template.FuncMap{
		"ext":          filepath.Ext,
		"sanitizePath": pathSanitizer,
		"nospace":      nospaceSanitizer,
	}

	for stName, stFunc := range util.FuncMap(o.namers) {
		funcs[stName] = stFunc
	}

	return funcs
}
