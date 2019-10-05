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

var (
	_ UserInterfacer = (*CLI)(nil) //Makes sure that CLI implements UserInterfacer
)

const (
	noOrEmptyValue = "<no value>"
	timeStampFmt   = time.RFC1123Z
	dateFmt        = "2006-01-02"

	emptyMediaType   = "empty"
	variousMediaType = "media"
)

//UserInterfacer represents any User Interface.
type UserInterfacer interface {
	//Printf displays a message to the user (has same behaviour than fmt.Printf)
	Printf(string, ...interface{})

	//PrettyPrint displays values from the provided map
	PrettyPrint(...map[string]interface{})

	//PrettyDiff displays povided maps, higlighting their differences
	PrettyDiff(map[string]interface{}, map[string]interface{})

	//Edit spawns an editor dialog to modified provided map
	Edit(map[string]interface{}) (map[string]interface{}, error)

	//Merge spawns a dialog to merge two maps into one
	Merge(map[string]interface{}, map[string]interface{}) (map[string]interface{}, error)
}

// CLIConfig describes configuration for User Interface
type CLIConfig struct {
	// Flag to switch between automatic or manual actions when editing or
	// merging records' attributes
	Auto bool

	// Command line to open a text editor
	//XXX: default to Getenv("EDITOR")?
	EditorCmd []string

	// Command line to open a text merger
	MergerCmd []string

	// Select the style of output to format answers (UIFormatters[UIFormatStyle])
	OutputFormat string

	// Templates to display information from the store.
	// Templates are organized by output style
	Formatters map[string]map[string]string
}

// ListStyles lists all available styles for printing records' details.
func (cfg *CLIConfig) ListStyles() (styles []string) {
	for k := range cfg.Formatters {
		styles = append(styles, k)
	}
	return
}

//CLI is a user interface built for the command line.
//If cfg.Auto flag is on, the returned User Interface will avoid any interaction
//with the user (like automatically merging metadata or skipping editing steps)
type CLI struct {
	editor []string
	merger []string

	style    style.Styler
	printers formatter.Formatters
}

//NewCLI creates a CLI User Interface
func NewCLI(cfg *CLIConfig) *CLI {
	ui := &CLI{
		style: style.NewColorterm(),

		printers: formatter.Formatters{
			formatter.DefaultFormatter: func(v interface{}) (string, error) {
				names := []string{}
				for _, m := range v.([]map[string]interface{}) {
					names = append(names, get(m, "Name"))
				}
				return strings.Join(names, "\n"), nil
			},
		},
	}

	if printers, exists := cfg.Formatters[cfg.OutputFormat]; exists {
		tmpl := template.New("UI").Funcs(map[string]interface{}{
			"metadata": ui.printMetadata,
			"table":    ui.listMediasByRows,
			//TODO(pimd): Add json, yaml...
		})

		for typ, txt := range printers {
			fmtFn := formatter.TemplateFormatter(tmpl.New(typ), txt)
			ui.printers.Register(typ, fmtFn)
		}
	}

	if !cfg.Auto {
		ui.editor, ui.merger = cfg.EditorCmd, cfg.MergerCmd
	}

	return ui
}

//Printf displays a message to the user (has same behaviour than fmt.Printf)
func (ui *CLI) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

//PrettyPrint shows in a pleasant manner a metadata set
func (ui *CLI) PrettyPrint(medias ...map[string]interface{}) {
	typ := mediasTypeOf(medias...)
	fmt.Println(ui.printers.MustFormatUsingType(typ, medias))
}

//PrettyDiff shows in a pleasant manner differences between two metadata sets
func (ui *CLI) PrettyDiff(mediaL, mediaR map[string]interface{}) {
	valL, valR := ui.format(mediaL), ui.format(mediaR)
	dT, dL, dR := text.ColorDiff.Anything(valL, valR)
	diffAsTable := text.NewTable().Col(dL, dT, dR).Draw()
	fmt.Println(diffAsTable)
}

//Edit fires-up a new editor to modif m
func (ui *CLI) Edit(m map[string]interface{}) (map[string]interface{}, error) {
	if len(ui.editor) > 0 {
		edited, err := input.EditAsJSON(m, ui.editor)
		return edited.(map[string]interface{}), err
	}

	return m, nil
}

//Merge fires-up a new editor to merge m and n
func (ui *CLI) Merge(m, n map[string]interface{}) (map[string]interface{}, error) {
	if len(ui.merger) > 0 {
		merged, _, err := input.MergeAsJSON(m, n, ui.merger)
		return merged.(map[string]interface{}), err
	}

	return merge(m, n)
}

func (ui *CLI) format(medias ...map[string]interface{}) string {
	typ := mediasTypeOf(medias...)
	return ui.printers.MustFormatUsingType(typ, medias)
}

func (ui *CLI) printMetadata(medias []map[string]interface{}, fields ...string) string {
	s := []string{}

	//keys is here essentially to list all fields in the same order
	//between medias
	keys := getCommonKeys(medias, fields...)

	for _, m := range medias {
		s = append(s, ui.listMediasByColumns([]map[string]interface{}{m}, keys...))
	}

	return strings.Join(s, "\n\n")
}

func (ui *CLI) listMediasByRows(medias []map[string]interface{}, fields ...string) string {
	table := text.NewTable()

	keys := getKeys(medias, fields...)
	table.Rows(styleSlice(ui.style.Bold, keys))

	for _, m := range medias {
		table.Rows(getValues(m, keys...))
	}

	return table.Draw()
}

func (ui *CLI) listMediasByColumns(medias []map[string]interface{}, fields ...string) string {
	table := text.NewTable()

	keys := getKeys(medias, fields...)
	table.Col(styleSlice(ui.style.Bold, keys))

	for _, m := range medias {
		table.Col(getValues(m, keys...))
	}

	return table.Draw()
}

func mediasTypeOf(medias ...map[string]interface{}) string {
	if len(medias) == 0 {
		return emptyMediaType
	}

	typ := formatter.TypeOf(medias[0])

	for _, m := range medias {
		if formatter.TypeOf(m) != typ {
			return variousMediaType
		}
	}
	return typ
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
			//So that possibly additional fields of a given maps can be extracted too
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

func hasKey(k string, maps ...map[string]interface{}) bool {
	for _, m := range maps {
		if _, exists := m[k]; !exists {
			return false
		}
	}
	return true
}

func isInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func styleSlice(fn func(string) string, s []string) (ss []string) {
	for _, txt := range s {
		ss = append(ss, fn(txt))
	}
	return
}

//merge completes m with n content with the following logic: values of m are
//copied, values of n that are not in m are added.
func merge(m, n map[string]interface{}) (map[string]interface{}, error) {
	merged := make(map[string]interface{})

	for k, v := range m {
		merged[k] = v
	}

	for k, v := range n {
		if _, exist := merged[k]; !exist {
			merged[k] = v
		}
	}

	return merged, nil
}
