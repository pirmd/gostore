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
	_ UserInterfacer = (*CLI)(nil)      //Makes sure that CLI implements UserInterfacer
	_ UserInterfacer = (*NoUserUI)(nil) //Makes sure that NoUserUI implements UserInterfacer
)

const (
	noOrEmptyValue = "<no value>"
	timeStampFmt   = time.RFC1123Z
	dateFmt        = "2006-01-02"

	diffMediaTypePreffix     = "diff_of_"
	emptyMediaType           = "empty"
	multipleMediaTypePreffix = "[]"
	genericMediaType         = "media"
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

//NewUI creates a new user interface.
//If auto flag is on, the returned User Interface will avoid any interaction
//with the user (like automatically merging metadata or skipping editing steps)
func NewUI(auto bool) UserInterfacer {
	//XXX: should be done at each NewCLI level (using cfg UIAuto flag ?)
	ui := newCLI()
	if auto {
		return &NoUserUI{UserInterfacer: ui}
	}
	return ui
}

//NoUserUI is a UserInterfacer that doesn't require any user interaction,
//notably it automatically merges metadata.
type NoUserUI struct{ UserInterfacer }

//Edit does nothing.
func (ui *NoUserUI) Edit(m map[string]interface{}) (map[string]interface{}, error) {
	return m, nil
}

//Merge completes m with n content with the following logic: values of m are
//copied, values of n that are not in m are added.
func (ui *NoUserUI) Merge(m, n map[string]interface{}) (map[string]interface{}, error) {
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

//CLI is a user interface built for the command line.
type CLI struct {
	editor []string
	merger []string

	printers formatter.Formatters
}

func newCLI() *CLI {
	ui := &CLI{
		editor: cfg.UIEditorCmd,
		merger: cfg.UIMergerCmd,

		printers: formatter.Formatters{
			formatter.DefaultFormatter: formatter.JSONFormatter,
		},
	}

	tmpl := template.New("ui").Funcs(map[string]interface{}{
		"showMetadata": showMetadata,
		"listMedia":    listByRows,
		"diff":         diff,
		"diffMedias":   diffMedias,
	})

	for typ, txt := range cfg.UIFormatters[cfg.UIFormatStyle] {
		fmtFn := formatter.TemplateFormatter(tmpl.New(typ), txt)
		ui.printers.Register(typ, fmtFn)
	}

	for typ, txt := range cfg.UIDiffers {
		diffType := diffMediaTypePreffix + typ
		fmtFn := formatter.TemplateFormatter(tmpl.New(diffType), txt)
		ui.printers.Register(diffType, fmtFn)
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
	typ := diffMediaTypePreffix + mediasTypeOf(mediaL, mediaR)
	fmt.Println(ui.printers.MustFormatUsingType(typ, struct{ L, R map[string]interface{} }{mediaL, mediaR}))
}

//Edit fires-up a new editor to modif v
func (ui *CLI) Edit(m map[string]interface{}) (map[string]interface{}, error) {
	edited, err := input.EditAsJSON(m, ui.editor)
	return edited.(map[string]interface{}), err
}

//Merge fires-up a new editor to merge v and w
func (ui *CLI) Merge(m, n map[string]interface{}) (map[string]interface{}, error) {
	merged, _, err := input.MergeAsJSON(m, n, ui.merger)
	return merged.(map[string]interface{}), err
}

func mediasTypeOf(medias ...map[string]interface{}) string {
	if len(medias) == 0 {
		return emptyMediaType
	}

	typ := formatter.TypeOf(medias[0])

	for _, m := range medias {
		if formatter.TypeOf(m) != typ {
			// provided medias list is of various type, return generic name
			return multipleMediaTypePreffix + genericMediaType
		}
	}
	return multipleMediaTypePreffix + typ
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
