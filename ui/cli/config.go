package cli

import (
	"fmt"
	"text/template"

	"github.com/kballard/go-shellquote"

	"github.com/pirmd/gostore/ui/formatter"
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

// NewFromConfig creates a CLI User Interface from a given Config
func NewFromConfig(cfg *Config) (*CLI, error) {
	ui := New()

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
