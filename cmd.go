package main

import (
	"fmt"

	"github.com/pirmd/cli/app"
)

var (
	gostoreApp = app.New("gostore", "A command-line minimalist media collection manager.")
)

func init() {
	gostoreApp.NewBoolFlagToVar(&cfg.LogDebug, "debug", "Show debug information.")
	gostoreApp.NewStringFlagToVar(&cfg.StoreRoot, "root", "Path to the root of the collection.")
	gostoreApp.NewBoolFlagToVar(&cfg.StoreReadOnly, "pretend", "Simulate operation on the database (actually retrieveing metadata without inserting them into the store).")
	gostoreApp.NewBoolFlagToVar(&cfg.UIAuto, "auto", "Automatically fecth metadata before importing them in the collection.")
	gostoreApp.NewEnumFlagToVar(&cfg.UIFormatStyle, "style", "Style for printing records' details.", getUIFormatStyles(cfg.UIFormatters))

	importCmd := gostoreApp.NewCommand("import", "Import a new media into the collection.")
	importMedia := importCmd.NewStringArg("media", "Media to import into the collection.")
	importCmd.Execute = func() error {
		gs := newGostore()

		if err := gs.Import(*importMedia); err != nil {
			return fmt.Errorf("importing '%s' failed: %s", *importMedia, err)
		}
		return nil
	}

	infoCmd := gostoreApp.NewCommand("info", "Retrieve information about any collection's record.")
	infoFromFile := infoCmd.NewBoolFlag("from-file", "Retrieve information from media file rather than from database.")
	infoKey := infoCmd.NewStringArg("name", "Name of the record to get information about.")
	infoCmd.Execute = func() error {
		gs := newGostore()

		if err := gs.Info(*infoKey, *infoFromFile); err != nil {
			return fmt.Errorf("getting information about '%s' failed: %s", *infoKey, err)
		}
		return nil
	}

	listCmd := gostoreApp.NewCommand("list", "List the collection's records matching the given pattern. If no pattern is provied, list all records of the collection.")
	listPattern := listCmd.NewStringArg("pattern", "Pattern to match records against. Pattern follows blevesearch query language (https://blevesearch.com/docs/Query-String-Query/).")
	listCmd.CanRunWithoutArg = true
	listCmd.Execute = func() error {
		gs := newGostore()

		if *listPattern != "" && *listPattern != "*" {
			if err := gs.ListAll(); err != nil {
				return fmt.Errorf("listing collection's content failed: %s", err)
			}
			return nil
		}

		if err := gs.Search(*listPattern); err != nil {
			return fmt.Errorf("listing records matching '%s' failed: %s", *listPattern, err)
		}
		return nil
	}

	editCmd := gostoreApp.NewCommand("edit", "Edit an existing record from the collection.")
	editKey := editCmd.NewStringArg("name", "Name of the record to edit.")
	editCmd.Execute = func() error {
		gs := newGostore()

		if err := gs.Edit(*editKey); err != nil {
			return fmt.Errorf("editing '%s' failed: %s", *editKey, err)
		}
		return nil
	}

	delCmd := gostoreApp.NewCommand("delete", "Delete a record from the collection.")
	delKey := delCmd.NewStringArg("name", "Name of the record to delete.")
	delCmd.Execute = func() error {
		gs := newGostore()

		if err := gs.Delete(*delKey); err != nil {
			return fmt.Errorf("deleting '%s' failed: %s", *delKey, err)
		}
		return nil
	}

	exportCmd := gostoreApp.NewCommand("export", "Copy a record's media file from the collection to the given destination.")
	exportKey := exportCmd.NewStringArg("name", "Name of the record to export.")
	exportDst := exportCmd.NewStringArg("dst", "Destination folder where the record needs to be exported.")
	exportCmd.Execute = func() error {
		gs := newGostore()

		if err := gs.Export(*exportKey, *exportDst); err != nil {
			return fmt.Errorf("exporting '%s' to '%s' failed: %s", *exportKey, *exportDst, err)
		}
		return nil
	}

	checkCmd := gostoreApp.NewCommand("check", "Verify collection's consistency and repairs or reports found inconsistencies.")
	checkCmd.Execute = func() error {
		gs := newGostore()

		if err := gs.CheckAndRepair(); err != nil {
			return fmt.Errorf("checking collection failed: %s", err)
		}
		return nil
	}
}
