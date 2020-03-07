package main

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/pirmd/clapp"
)

func newApp(cfg *Config) *clapp.Command {
	cmd := &clapp.Command{
		Name:        "gostore",
		Usage:       "A command-line minimalist media collection manager.",
		Description: "gostore is a command line tool aiming at providing facilities to manage one or more collections of media files, keeping track of their metadata and/or additional information the user wants to record.",

		Config: &clapp.Config{
			Unmarshaller: yaml.Unmarshal,
			Files:        clapp.DefaultConfigFiles("config.yaml"),
			Var:          cfg,
		},

		ShowHelp:    clapp.ShowUsage,
		ShowVersion: clapp.ShowVersion,
	}

	cmd.Flags = clapp.Flags{
		{
			Name:  "debug",
			Usage: "Show debug information.",
			Var:   &cfg.ShowLog,
		},

		{
			Name:  "root",
			Usage: "Path to the root of the collection.",
			Var:   &cfg.Store.Root,
		},

		{
			Name:  "pretend",
			Usage: "Operations that modify the collection are simulated.",
			Var:   &cfg.Store.ReadOnly,
		},

		{
			Name:  "auto",
			Usage: "Perform operations without manual interaction from the user.",
			Var:   &cfg.UI.Auto,
		},

		{
			Name:  "style",
			Usage: "Style for printing records' details. Available styles are defined in the configuration file.",
			Var:   &cfg.UI.OutputFormat,
		},
	}

	var mediaPath string
	cmd.SubCommands.Add(&clapp.Command{
		Name:  "import",
		Usage: "Import a new media into the collection.",
		Args: clapp.Args{
			{
				Name:  "media",
				Usage: "Media to import into the collection.",
				Var:   &mediaPath,
			},
		},
		Execute: func() error {
			gs, err := newGostore(cfg)
			if err != nil {
				return err
			}

			if err := gs.Import(mediaPath); err != nil {
				return fmt.Errorf("importing '%s' failed: %s", mediaPath, err)
			}
			return nil
		},
	})

	var recordID string
	var readInfoFromMediaFile bool
	cmd.SubCommands.Add(&clapp.Command{
		Name:  "info",
		Usage: "Retrieve information about any collection's record.",

		Flags: clapp.Flags{
			{
				Name:  "from-file",
				Usage: "Read information from media file rather than from the collection",
				Var:   &readInfoFromMediaFile,
			},
		},

		Args: clapp.Args{
			{
				Name:  "name",
				Usage: "Name of the record to get information about.",
				Var:   &recordID,
			},
		},

		Execute: func() error {
			gs, err := newGostore(cfg)
			if err != nil {
				return err
			}

			if err := gs.Info(recordID, readInfoFromMediaFile); err != nil {
				return fmt.Errorf("getting information about '%s' failed: %s", recordID, err)
			}
			return nil
		},
	})

	searchPattern := "*"
	cmd.SubCommands.Add(&clapp.Command{
		Name: "list",

		Usage: "Lists the collection's records matching the given pattern. If no pattern is provied, list all records of the collection.",

		Args: clapp.Args{
			{
				Name:     "pattern",
				Usage:    "Pattern to match records against. Pattern follows blevesearch query language (https://blevesearch.com/docs/Query-String-Query/).",
				Var:      &searchPattern,
				Optional: true,
			},
		},

		Execute: func() error {
			gs, err := newGostore(cfg)
			if err != nil {
				return err
			}

			if searchPattern != "*" {
				if err := gs.ListAll(); err != nil {
					return fmt.Errorf("listing collection's content failed: %s", err)
				}
				return nil
			}

			if err := gs.Search(searchPattern); err != nil {
				return fmt.Errorf("listing records matching '%s' failed: %s", searchPattern, err)
			}
			return nil
		},
	})

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "edit",
		Usage: "Edit an existing record from the collection using user defined's editor. If flag '--auto' is used, edition is skipped and nothing happens.",

		Args: clapp.Args{
			{
				Name:  "name",
				Usage: "Name of the record to edit.",
				Var:   &recordID,
			},
		},

		Execute: func() error {
			gs, err := newGostore(cfg)
			if err != nil {
				return err
			}

			if err := gs.Edit(recordID); err != nil {
				return fmt.Errorf("editing '%s' failed: %s", recordID, err)
			}
			return nil
		},
	})

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "delete",
		Usage: "Delete an existing record from the collection.",

		Args: clapp.Args{
			{
				Name:  "name",
				Usage: "Name of the record to delete.",
				Var:   &recordID,
			},
		},

		Execute: func() error {
			gs, err := newGostore(cfg)
			if err != nil {
				return err
			}

			if err := gs.Delete(recordID); err != nil {
				return fmt.Errorf("deleting '%s' failed: %s", recordID, err)
			}
			return nil
		},
	})

	var dstFolder string
	cmd.SubCommands.Add(&clapp.Command{
		Name:  "export",
		Usage: "Copy a record's media file from the collection to the given destination.",

		Args: clapp.Args{
			{
				Name:  "name",
				Usage: "Name of the record to export.",
				Var:   &recordID,
			},
			{
				Name:     "dst",
				Usage:    "Destination folder where the record needs to be exported to. Default to current working directory.",
				Var:      &dstFolder,
				Optional: true,
			},
		},

		Execute: func() error {
			gs, err := newGostore(cfg)
			if err != nil {
				return err
			}

			if err := gs.Export(recordID, dstFolder); err != nil {
				return fmt.Errorf("exporting '%s' to '%s' failed: %s", recordID, dstFolder, err)
			}
			return nil
		},
	})

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "check",
		Usage: "Verify collection's consistency and repairs or reports found inconsistencies.",
		Execute: func() error {
			gs, err := newGostore(cfg)
			if err != nil {
				return err
			}

			if err := gs.CheckAndRepair(); err != nil {
				return fmt.Errorf("checking collection failed: %s", err)
			}
			return nil
		},
	})

	return cmd
}
