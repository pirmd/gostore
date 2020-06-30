package cli

import (
	"encoding/json"
	"text/template"

	"github.com/pirmd/gostore/util"

	"github.com/pirmd/style"
	"github.com/pirmd/text"
)

func (ui *CLI) funcmap() template.FuncMap {
	funcs := template.FuncMap{
		"getAll": getAllMetadata,
		"get":    getMetadata,
		"bycol":  tableCol,
		"byrow":  tableRow,

		"extend": func(m map[string]interface{}, key, val string) string {
			m[key] = val
			return ""
		},

		"json": func(v interface{}) (string, error) {
			output, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(output), nil
		},
		"jsonForHuman": func(v interface{}) (string, error) {
			output, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return "", err
			}
			return string(output), nil
		},
	}

	for stName, stFunc := range util.FuncMap(ui.printers) {
		funcs[stName] = stFunc
	}

	for stName, stFunc := range style.FuncMap(ui.style) {
		funcs[stName] = styleTable(stFunc.(func(string) string))
	}

	return funcs
}

func getAllMetadata(maps []map[string]interface{}, fields ...string) [][]string {
	return map2kv(maps, fields...).KV()
}

func getMetadata(m map[string]interface{}, fields ...string) [][]string {
	maps := []map[string]interface{}{m}
	return getAllMetadata(maps, fields...)
}

func tableCol(col [][]string) string {
	return text.NewTable().Col(col...).Draw()
}

func tableRow(rows [][]string) string {
	return text.NewTable().Rows(rows...).Draw()
}

func styleTable(fn func(string) string) func(tab [][]string, idx ...int) [][]string {
	return func(tab [][]string, idx ...int) [][]string {
		if len(tab) == 0 {
			return [][]string{}
		}

		if len(idx) == 0 {
			idx = []int{0}
		}

		stab := make([][]string, len(tab))
		for i := range tab {
			stab[i] = make([]string, len(tab[i]))
			copy(stab[i], tab[i])
		}

		for _, i := range idx {
			for j, s := range stab[i] {
				stab[i][j] = fn(s)
			}
		}

		return stab
	}
}
