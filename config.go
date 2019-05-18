package main

// This module implements a more than basic configuration system

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pirmd/cli/style"
	"github.com/pirmd/gostore/processing"
)

var (
	cfg = struct {
		// Configuration for storage
		StoreRoot              string            // Path to the datastore
		StoreNamingSchemes     map[string]string // Template to save a book to disk
		StoreRecordProcessings []string          // List of processings to be applied when importing or updating a record

		// UI preference
		UIEditorCmd   []string                     // Command line to open a text editor
		UIMergerCmd   []string                     // Command line to open a text merger
		UIFormatStyle string                       // Select the style of output to format answers (UIFormatters[UIFormatStyle])
		UIFormatters  map[string]map[string]string // Templates to display information from the store.
		// Templates are organized by output style
		UIDiffers map[string]string // Templates to display differences between two records or two records versions

		// Logging
		LogVerbose bool
		LogDebug   bool
	}{
		StoreRoot: ".", //Current working directory
		StoreNamingSchemes: map[string]string{
			"_default": "{{if not .Authors}}unknown{{else}}{{with index .Authors 0}}{{.}}{{end}}{{end}} - {{.Title}}",
		},

		StoreRecordProcessings: []string{
			"rename",
		},

		UIFormatStyle: "name",

		UIFormatters: map[string]map[string]string{
			"name": map[string]string{
				"_default": "{{range $i, $m := .}}{{if $i}}\n{{end}}{{$m.Name}}{{end}}",
			},

			"list": map[string]string{
				"_default": `{{ listMedia . "Name" "Title" "Authors" }}`,
				"[]epub":   `{{ listMedia . "Name" "Title" "Serie" "SeriePosition" "Authors" }}`,
				"empty":    `no match`,
			},

			"full": map[string]string{
				"_default": `{{ showMetadata . "Title" "*" "Name" | colorMissing }}`,
				"[]epub":   `{{ showMetadata . "Name" "Title" "Serie" "SeriePosition" "Authors" "Description" "*" "Type" "CreatedAt" "UpdatedAt" | colorMissing }}`,
				"empty":    `no match`,
			},

			"json": map[string]string{},
		},

		UIDiffers: map[string]string{
			"_default": `{{ diffMedias .L .R "Title" "*" "Name" }}`,
			"[]epub":   `{{ diffMedias .L .R "Title" "Serie" "SeriePosition" "Authors" "Description" "*" "Type" }}`,
		},

		UIEditorCmd: []string{os.Getenv("EDITOR")},
		UIMergerCmd: []string{"vimdiff"},
	}

	debugger = log.New(ioutil.Discard, "DEBUG ", log.Ltime|log.Lshortfile)
)

func configure() {
	for typ, txt := range cfg.StoreNamingSchemes {
		processing.AddNamer(typ, txt)
	}

	processing.RecordProcessings = cfg.StoreRecordProcessings

	if cfg.LogDebug {
		debugger.SetOutput(os.Stderr)
	}

	style.CurrentStyler = style.ColorTerm

	for typ, txt := range cfg.UIFormatters[cfg.UIFormatStyle] {
		AddPrettyPrinter(typ, txt)
	}

	for typ, txt := range cfg.UIDiffers {
		AddPrettyDiffer(typ, txt)
	}
}

func getUIFormatStyles(m map[string]map[string]string) (styles []string) {
	for k, _ := range m {
		styles = append(styles, k)
	}
	return
}
