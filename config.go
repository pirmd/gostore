package main

// This module implements a more than basic configuration system

// XXX: restructure like module NewFromYAML + newGostore, same for UI, LOG and Cie. Try to generalize module format
// XXX: create a ConfigLoader parameter that can be JSON, TOML of YAML or whatever

import (
	"os"
)

// Configuration for gostore
type config struct {
	ShowLog       bool
	Store         *storeConfig
	UI            *CLIConfig
	ImportModules []string // List of processings to be applied when importing a record
	UpdateModules []string // List of processings to be applied when updating a record
	Modules       map[string][]byte
}

// Configuration for storage
type storeConfig struct {
	//Root contains the path to the datastore
	Root string
	//ReadOnly is the flag to switch the store into read only operation mode
	//XXX it is not implemented
	ReadOnly bool
}

var (
	//XXX -> config.yaml.example
	cfg = &config{
		UI: &CLIConfig{
			Formatters: map[string]map[string]string{
				"list": {
					"_default": `{{ table . "Name" "Title" "Authors" }}`,
					"epub":     `{{ table . "Name" "Title" "SubTitle" "Serie" "SeriePosition" "Authors" }}`,
				},

				"full": {
					"_default": `{{ metadata . "Name" "Title" "*" "CreatedAt" "UpdatedAt"}}`,
					"epub":     `{{ metadata . "Name" "Title" "SubTitle" "Serie" "SeriePosition" "Authors" "Description" "*" "Type" "CreatedAt" "UpdatedAt" }}`,
				},
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

		Modules: map[string][]byte{
			"organizer": []byte(`
NamingSchemes:
    _default: {{if not .Authors}}unknown{{else}}{{with index .Authors 0}}{{.}}{{end}}{{end}} - {{.Title}}
Sanitizer    : detox
            `),

			"dehtmlizer": []byte(`
Fields2Clean:
    - Description
OutputStyle : markdown
            `),
		},
	}
)
