package main

// This module implements a more than basic configuration system

import (
	"io/ioutil"
	"log"
	"os"

	//XXX: It i snot convenient to have to imports all modules
	"github.com/pirmd/gostore/modules/dehtmlizer"
	"github.com/pirmd/gostore/modules/organizer"
)

type gostoreConfig struct {
	Store         *storeConfig
	Log           *logConfig
	UI            *uiConfig
	ImportModules []string // List of processings to be applied when importing a record
	UpdateModules []string // List of processings to be applied when updating a record
	Modules       map[string]interface{}

	// Logging
	LogVerbose bool
	LogDebug   bool
}

// Configuration for storage
type storeConfig struct {
	//Root contains the path to the datastore
	Root string
	//ReadOnly is the flag to switch the store into read only operation mode
	//XXX it is not implemented
	ReadOnly bool
}

// Configuration for User Interface
type uiConfig struct {
	// Flag to switch between automatic or manual actions when editing or
	// merging records' attributes
	Auto bool

	// Command line to open a text editor
	EditorCmd []string

	// Command line to open a text merger
	MergerCmd []string

	// Select the style of output to format answers (UIFormatters[UIFormatStyle])
	FormatStyle string

	// Templates to display information from the store.
	// Templates are organized by output style
	Formatters map[string]map[string]string
	//XXX: simplify fomrat for style output (see go list -f FORMAT)

	// Templates to display differences between two records or two records versions
	Differs map[string]string
}

// ListStyles lists all available styles for printing records' details.
func (cfg *uiConfig) ListStyles() (styles []string) {
	for k := range cfg.Formatters {
		styles = append(styles, k)
	}
	return
}

// Logger configuration
type logConfig struct {
	Debug bool
}

// Logger creates a new logger corresponding to the logConfig parameters
func (cfg *logConfig) Logger() *log.Logger {
	//XXX: change name to Logger instead of debugger
	//XXX: simplify debugger using standard log feature (?)
	//XXX: improve logger module (more customization maybe)
	//XXX: change logger into a module (?)
	if cfg.Debug {
		return log.New(os.Stderr, "DEBUG ", log.Ltime|log.Lshortfile)
	}

	return log.New(ioutil.Discard, "DEBUG ", log.Ltime|log.Lshortfile)
}

var (
	cfg = &gostoreConfig{
		Store: &storeConfig{Root: "."}, //Current working directory

		Log: &logConfig{},

		UI: &uiConfig{
			FormatStyle: "name",

			Formatters: map[string]map[string]string{
				"name": {
					"_default": "{{range $i, $m := .}}{{if $i}}\n{{end}}{{$m.Name}}{{end}}",
				},

				"list": {
					"_default": `{{ listMedia . "Name" "Title" "Authors" }}`,
					"[]epub":   `{{ listMedia . "Name" "Title" "SubTitle" "Serie" "SeriePosition" "Authors" }}`,
					"empty":    `no match`,
				},

				"full": {
					"_default": `{{ showMetadata . "Title" "*" "Name" }}`,
					"[]epub":   `{{ showMetadata . "Name" "Title" "SubTitle" "Serie" "SeriePosition" "Authors" "Description" "*" "Type" "CreatedAt" "UpdatedAt" }}`,
					"empty":    `no match`,
				},

				"json": {},
			},

			//XXX: is it really needed? can live on top of Formatter?
			Differs: map[string]string{
				"_default": `{{ diffMedias .L .R "Title" "*" "Name" }}`,
				"[]epub":   `{{ diffMedias .L .R "Title" "SubTitle" "Serie" "SeriePosition" "Authors" "Description" "*" "Type" }}`,
			},

			EditorCmd: []string{os.Getenv("EDITOR")},
			MergerCmd: []string{"vimdiff"},
		},

		ImportModules: []string{
			"organizer",
			"dehtmlizer",
		},

		UpdateModules: []string{
			"organizer",
			"dehtmlizer",
		},

		Modules: map[string]interface{}{
			"organizer": &organizer.Config{
				NamingSchemes: map[string]string{
					"_default": "{{if not .Authors}}unknown{{else}}{{with index .Authors 0}}{{.}}{{end}}{{end}} - {{.Title}}",
				},
				Sanitizer: "detox",
			},

			"dehtmlizer": &dehtmlizer.Config{
				Fields2Clean: []string{"Description"},
				OutputStyle:  "markdown",
			},
		},
	}
)
