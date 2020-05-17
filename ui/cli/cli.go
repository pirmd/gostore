package cli

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pirmd/style"
	"github.com/pirmd/text/diff"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/ui"
)

const (
	// DefaultFormatter is the name of the fallback Formatter
	DefaultFormatter = media.DefaultType
)

var (
	_ ui.UserInterfacer = (*CLI)(nil) //Makes sure that CLI implements UserInterfacer

)

// CLI is a user interface built for the command line.
// If cfg.Auto flag is on, the returned User Interface will avoid any interaction
// with the user (like automatically merging metadata or skipping editing steps)
type CLI struct {
	editor []string
	merger []string

	style    style.Styler
	printers *template.Template
}

// New creates CLI User Interface with default values
func New() *CLI {
	ui := &CLI{
		style:    style.NewColorterm(),
		printers: template.New("pprinter"),
	}
	ui.printers.Funcs(ui.funcmap())

	return ui
}

// Printf displays a message to the user (has same behaviour than fmt.Printf)
func (ui *CLI) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// PrettyPrint shows in a pleasant manner a metadata set
func (ui *CLI) PrettyPrint(medias ...map[string]interface{}) {
	fmt.Print(ui.print(medias...))
}

// PrettyDiff shows in a pleasant manner differences between two metadata sets
func (ui *CLI) PrettyDiff(mediaL, mediaR map[string]interface{}) {
	deltaL := make(map[string]interface{})

	allkeys := getKeys([]map[string]interface{}{mediaL, mediaR}, "?*")
	for _, key := range allkeys {
		dL, _, _, _ := diff.Patience(get(mediaL, key), get(mediaR, key), diff.ByLines, diff.ByWords).PrettyPrint(diff.WithColor)
		deltaL[key[1:]] = strings.Join(dL, "")
	}

	ui.PrettyPrint(deltaL)
}

// Edit fires-up a new editor to modify a map
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

func (ui *CLI) print(medias ...map[string]interface{}) string {
	t := ui.printerFor(medias...)
	if t == nil {
		return fmt.Sprintf("%+v", medias)
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, medias); err != nil {
		return fmt.Sprintf("!Err(%s)", err)
	}
	return buf.String()
}

func (ui *CLI) printerFor(medias ...map[string]interface{}) *template.Template {
	typ := typeOf(medias...)

	if tmpl := ui.printers.Lookup(typ); tmpl != nil {
		return tmpl
	}

	if tmpl := ui.printers.Lookup(filepath.Base(typ)); tmpl != nil {
		return tmpl
	}

	if tmpl := ui.printers.Lookup(filepath.Dir(typ)); tmpl != nil {
		return tmpl
	}

	return ui.printers.Lookup(DefaultFormatter)
}

// typeOf returns a common type for a collection of maps. If maps are not of
// the same type, it returns media.DefaultType
func typeOf(maps ...map[string]interface{}) string {
	if len(maps) == 0 {
		return media.DefaultType
	}

	var typ string
	for i, m := range maps {
		if i == 0 {
			typ = media.TypeOf(m)
			continue
		}

		if media.TypeOf(m) != typ {
			return media.DefaultType
		}
	}

	return typ
}
