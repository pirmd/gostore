package main

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/pirmd/cli/formatter"
	"github.com/pirmd/cli/input"
	"github.com/pirmd/cli/style"
	"github.com/pirmd/cli/style/text"
)

const (
	noOrEmptyValue = "<no value>"
	timeStampFmt   = time.RFC1123Z
	dateFmt        = "2006-01-02"
)

var (
	printerTmpl = template.New("ui").Funcs(map[string]interface{}{
		"showMetadata": showMetadata,
		"listMedia":    listByRows,

		"values": getValues,
		"keys":   getKeys,
		"slice":  func(s ...string) []string { return s },
		"newRow": func(r []string) [][]string { return append([][]string{}, r) },
		"addRow": func(a [][]string, s []string) [][]string { return append(a, s) },
		"table":  func(r [][]string) string { return text.NewTable().Rows(r...).String() },
	})

	pprinters = formatter.Formatters{
		formatter.DefaultFormatter: formatter.JSONFormatter,
	}

	differTmpl = template.New("ui").Funcs(map[string]interface{}{
		"diff":       diff,
		"diffMedias": diffMedias,
	})

	pdiffers = formatter.Formatters{
		formatter.DefaultFormatter: formatter.JSONFormatter,
	}
)

//AddPrettyPrinter registers a new pretty printer
func AddPrettyPrinter(name string, text string) {
	pprinters.Register(name, formatter.TemplateFormatter(printerTmpl.New(name), text))
}

//AddPrettyDiffer registers a new pretty differ
func AddPrettyDiffer(name string, text string) {
	pdiffers.Register(name, formatter.TemplateFormatter(differTmpl.New(name), text))
}

//Edit fires-up a new editor to modif v
func Edit(v interface{}) (interface{}, error) {
	return input.EditAsJSON(v, cfg.UIEditorCmd)
}

//Merge fires-up a new editor to merge v and w
func Merge(v, w interface{}) (interface{}, interface{}, error) {
	return input.MergeAsJSON(v, w, cfg.UIMergerCmd)
}

//PrettyDiff shows in a pleasant manner differences between two metadata sets
func PrettyDiff(mediaL, mediaR map[string]interface{}, fields ...string) {
	fmt.Println(pdiffers.MustFormatUsingType(mediasTypeOf([]map[string]interface{}{mediaL, mediaR}), struct{ L, R map[string]interface{} }{mediaL, mediaR}))
}

//PrettyPrint shows in a pleasant manner a metadata set
func PrettyPrint(medias ...map[string]interface{}) {
	fmt.Println(pprinters.MustFormatUsingType(mediasTypeOf(medias), medias))
}

func mediasTypeOf(medias []map[string]interface{}) string {
	if len(medias) == 0 {
		return "empty"
	}

	typ := formatter.TypeOf(medias[0])
	for _, m := range medias {
		if formatter.TypeOf(m) != typ {
			return "[]media"
		}
	}
	return "[]" + typ
}

func styleSlice(s []string, fn func(string) string) (ss []string) {
	for _, txt := range s {
		ss = append(ss, fn(txt))
	}
	return
}

func showMetadata(medias []map[string]interface{}, fields ...string) string {
	s := []string{}

	//keys is here essentially to list all fields in the same order
	//between medias
	keys := getCommonKeys(medias, fields...)

	for _, m := range medias {
		s = append(s, listByColumns([]map[string]interface{}{m}, keys...))
	}

	return strings.Join(s, "\n\n")
}

func listByRows(medias []map[string]interface{}, fields ...string) string {
	table := text.NewTable()

	keys := getKeys(medias, fields...)
	table.Rows(styleSlice(keys, style.CurrentStyler.Bold))

	for _, m := range medias {
		table.Rows(getValues(m, keys...))
	}

	return table.Draw()
}

func listByColumns(medias []map[string]interface{}, fields ...string) string {
	table := text.NewTable()

	keys := getKeys(medias, fields...)
	table.Col(styleSlice(keys, style.CurrentStyler.Bold))

	for _, m := range medias {
		table.Col(getValues(m, keys...))
	}

	return table.Draw()
}

func diffMedias(mediaL, mediaR map[string]interface{}, fields ...string) string {
	keys := getKeys([]map[string]interface{}{mediaL, mediaR}, fields...)
	valL, valR := getValues(mediaL, keys...), getValues(mediaR, keys...)

	dT, dL, dR := text.ColorDiff.Slices(valL, valR)

	return text.NewTable().Col(styleSlice(keys, style.CurrentStyler.Bold), dL, dT, dR).Draw()
}

func diff(l, r interface{}) string {
	dT, dL, dR := text.ColorDiff.Anything(l, r)
	return text.NewTable().Col(dL, dT, dR).Draw()
}

func getKeys(maps []map[string]interface{}, fields ...string) (keys []string) {
	if len(fields) == 0 {
		fields = []string{"*"}
	}

	for _, f := range fields {
		switch f {
		case "*":
			for _, m := range maps {
				for k := range m {
					if !isInSlice(k, fields) && !isInSlice(k, keys) {
						keys = append(keys, k)
					}
				}
			}

		default:
			keys = append(keys, f)
		}
	}

	return
}

func getCommonKeys(maps []map[string]interface{}, fields ...string) (keys []string) {
	if len(fields) == 0 {
		fields = []string{"*"}
	}

	for _, f := range fields {
		switch f {
		case "*":
			//We only keep keys that are present in all maps, then we pass the "*"
			//So that possiblt additional fields of a given maps can be extracted too
			if len(maps) > 0 {
				for k := range maps[0] {
					if !isInSlice(k, fields) && !isInSlice(k, keys) &&
						hasKey(k, maps[1:]...) {
						keys = append(keys, k)
					}
				}
			}
			keys = append(keys, f)

		default:
			keys = append(keys, f)
		}
	}

	return
}

func getValues(m map[string]interface{}, fields ...string) (values []string) {
	for _, f := range fields {
		switch f {
		case "*":
			for k := range m {
				if !isInSlice(k, fields) {
					values = append(values, get(m, k))
				}
			}

		default:
			values = append(values, get(m, f))
		}
	}
	return
}

func get(m map[string]interface{}, k string) string {
	if v, exists := m[k]; exists {
		if t, ok := v.(time.Time); ok {
			//Only a date
			if strings.HasSuffix(k, "Date") {
				return t.Format(dateFmt)
			}
			//Stamp
			return t.Format(timeStampFmt)
		}

		if value := fmt.Sprintf("%v", v); value != "" {
			return value
		}
	}

	return noOrEmptyValue
}

func isInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func hasKey(k string, maps ...map[string]interface{}) bool {
	for _, m := range maps {
		if _, exists := m[k]; !exists {
			return false
		}
	}
	return true
}
