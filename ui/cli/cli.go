package cli

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/kballard/go-shellquote"

	"github.com/pirmd/style"
	"github.com/pirmd/text"
	"github.com/pirmd/text/diff"

	"github.com/pirmd/gostore/ui"
	"github.com/pirmd/gostore/ui/formatter"
)

var (
	_ ui.UserInterfacer = (*CLI)(nil) //Makes sure that CLI implements UserInterfacer
)

const (
	noOrEmptyValue = "<no value>"
	timeStampFmt   = time.RFC1123Z
	dateFmt        = "2006-01-02"

	emptyMediaType   = "empty"
	variousMediaType = "media"
)

// Config describes configuration for User Interface
type Config struct {
	// Flag to switch between automatic or manual actions when editing or
	// merging records' attributes
	Auto bool

	// Command line to open a text editor
	EditorCmd string

	// Command line to open a text merger
	MergerCmd string

	// Select the style of output to format answers (UIFormatters[UIFormatStyle])
	OutputFormat string

	// Templates to display information from the store.
	// Templates are organized by output style
	Formatters map[string]map[string]string
}

// ListStyles lists all available styles for printing records' details.
func (cfg *Config) ListStyles() (styles []string) {
	for k := range cfg.Formatters {
		styles = append(styles, k)
	}
	return
}

// CLI is a user interface built for the command line.
// If cfg.Auto flag is on, the returned User Interface will avoid any interaction
// with the user (like automatically merging metadata or skipping editing steps)
type CLI struct {
	editor []string
	merger []string

	style    style.Styler
	printers formatter.Formatters
}

// New creates a CLI User Interface
func New(cfg *Config) (*CLI, error) {
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
		var err error

		if ui.editor, err = shellquote.Split(cfg.EditorCmd); err != nil {
			return nil, fmt.Errorf("CLI config: parsing EditorCmd failed: %v", err)
		}

		if ui.merger, err = shellquote.Split(cfg.MergerCmd); err != nil {
			return nil, fmt.Errorf("CLI config: parsing MergerCmd failed: %v", err)
		}
	}

	return ui, nil
}

// Printf displays a message to the user (has same behaviour than fmt.Printf)
func (ui *CLI) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// PrettyPrint shows in a pleasant manner a metadata set
func (ui *CLI) PrettyPrint(medias ...map[string]interface{}) {
	fmt.Println(ui.format(medias...))
}

// PrettyDiff shows in a pleasant manner differences between two metadata sets
func (ui *CLI) PrettyDiff(mediaL, mediaR map[string]interface{}) {
	valL, valR := ui.format(mediaL), ui.format(mediaR)
	dT, dL, dR, _ := diff.Patience(valL, valR, diff.ByLines, diff.ByWords).PrettyPrint(diff.WithColor, diff.WithoutMissingContent)
	diffAsTable := text.NewTable().Col(dL, dT, dR).Draw()
	fmt.Println(diffAsTable)
}

// Edit fires-up a new editor to modif m
func (ui *CLI) Edit(m map[string]interface{}) (map[string]interface{}, error) {
	if len(ui.editor) > 0 {
		edited, err := editAsJSON(m, ui.editor)
		return edited.(map[string]interface{}), err
	}

	return m, nil
}

// Merge fires-up a new editor to merge m and n
func (ui *CLI) Merge(m, n map[string]interface{}) (map[string]interface{}, error) {
	if len(ui.merger) > 0 {
		merged, _, err := mergeAsJSON(m, n, ui.merger)
		return merged.(map[string]interface{}), err
	}

	return mergeMaps(m, n)
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
		switch {
		case f == "*":
			var allkeys []string
			for _, m := range maps {
				for k := range m {
					if !isInSlice("!"+k, fields) &&
						!isInSlice(k, fields) &&
						!isInSlice(k, allkeys) {
						allkeys = append(allkeys, k)
					}
				}
			}
			sort.Strings(allkeys)
			keys = append(keys, allkeys...)

		case f[0] == '!':
			//ignore this field

		default:
			keys = append(keys, f)
		}
	}

	return
}

func getCommonKeys(maps []map[string]interface{}, fields ...string) (keys []string) {
	if len(maps) == 0 {
		return
	}

	allKeys := getKeys([]map[string]interface{}{maps[0]}, fields...)

	for _, k := range allKeys {
		if hasKey(k, maps[1:]...) {
			keys = append(keys, k)
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
		if s == item {
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

// mergeMaps completes m with n content with the following logic: values of m are
// copied, values of n that are not in m are added.
func mergeMaps(m, n map[string]interface{}) (map[string]interface{}, error) {
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
