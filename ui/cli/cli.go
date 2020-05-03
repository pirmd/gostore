package cli

import (
	"fmt"
	"strings"

	"github.com/pirmd/style"
	"github.com/pirmd/text"
	"github.com/pirmd/text/diff"

	"github.com/pirmd/gostore/ui"
	"github.com/pirmd/gostore/ui/formatter"
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
	printers formatter.Formatters
}

// New creates CLI User Interface with default values
func New() *CLI {
	return &CLI{
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
}

// Printf displays a message to the user (has same behaviour than fmt.Printf)
func (ui *CLI) Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// PrettyPrint shows in a pleasant manner a metadata set
func (ui *CLI) PrettyPrint(medias ...map[string]interface{}) {
	fmt.Print(ui.format(medias...))
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
	typ := typeOf(medias...)
	return ui.printers.MustFormatUsingType(typ, medias)
}
