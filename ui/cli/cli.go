package cli

import (
	"fmt"
	"strings"
	"text/template"

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
		tmpl := template.New("UI").Funcs(ui.funcmap())

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
