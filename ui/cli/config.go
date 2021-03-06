package cli

import (
	"fmt"
)

// Config describes configuration for User Interface
type Config struct {
	// Auto is the flag that switches between automatic or manual actions when
	// editing or merging records' attributes
	Auto bool

	// EditorCmd contains the command line to open a text editor
	EditorCmd string

	// MergerCmd contains the command line to open a text merger
	MergerCmd string

	// OutputFormat selects the style of output to format records when printing
	// them.
	OutputFormat string

	// Formatters contains the list of templates to display information from
	// the store. Templates are organized by output style
	Formatters map[string]map[string]string
}

// NewConfig create a new Config
func NewConfig() *Config {
	return &Config{
		OutputFormat: "name",
		Formatters: map[string]map[string]string{
			"name": {
				DefaultFormatter: `{{ range $i, $r := . -}}
				{{- if $i }}{{ println }}{{ end -}}
				{{- .Name -}}
				{{- end -}}`,
			},
		},
	}
}

// Styles lists available styles for printing records.
func (cfg *Config) Styles() (styles []string) {
	for k := range cfg.Formatters {
		styles = append(styles, k)
	}
	return
}

// NewFromConfig creates a CLI User Interface from a given Config
func NewFromConfig(cfg *Config) (*CLI, error) {
	ui := New()

	printers, exists := cfg.Formatters[cfg.OutputFormat]
	if !exists {
		return nil, fmt.Errorf("CLI config: '%s': unknown formatter", cfg.OutputFormat)
	}

	for typ, txt := range printers {
		if _, err := ui.printers.New(typ).Parse(txt); err != nil {
			return nil, err
		}
	}

	if !cfg.Auto {
		ui.editor = cfg.EditorCmd
		ui.merger = cfg.MergerCmd
	}

	return ui, nil
}
