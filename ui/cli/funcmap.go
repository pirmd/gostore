package cli

import (
	"text/template"

	"github.com/pirmd/text"
)

func (ui *CLI) funcmap() template.FuncMap {
	return template.FuncMap{
		//TODO(pirmd): add json, yaml...
		//TODO(pirmd): document available formatting in the config.example
		//TODO(pirmd): modify github.com/pirmd/style to introduce ui.style.Funcmap

		"getAll": getAllMetadata,
		"get":    getMetadata,
		"bycol":  tableCol,
		"byrow":  tableRow,

		"upper":     styleTable(ui.style.Upper),
		"lower":     styleTable(ui.style.Lower),
		"titlecase": styleTable(ui.style.TitleCase),
		"black":     styleTable(ui.style.Black),
		"red":       styleTable(ui.style.Red),
		"green":     styleTable(ui.style.Green),
		"yellow":    styleTable(ui.style.Yellow),
		"blue":      styleTable(ui.style.Blue),
		"magenta":   styleTable(ui.style.Magenta),
		"cyan":      styleTable(ui.style.Cyan),
		"white":     styleTable(ui.style.White),
		"inverse":   styleTable(ui.style.Inverse),
		"bold":      styleTable(ui.style.Bold),
		"italic":    styleTable(ui.style.Italic),
		"underline": styleTable(ui.style.Underline),
		"crossout":  styleTable(ui.style.Crossout),
	}
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
