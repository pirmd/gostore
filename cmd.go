package main

import (
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
			Var:   &cfg.Store.Path,
		},

		{
			Name:  "pretend",
			Usage: "Operations that modify the collection are simulated.",
			Var:   &cfg.ReadOnly,
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

	var mediaPath []string
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
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.Import(mediaPath); err != nil {
				return err
			}
			return nil
		},
	})

	var recordID string

	var recordIDs []string
	cmd.SubCommands.Add(&clapp.Command{
		Name:  "list",
		Usage: "List and retrieve information about collection's records. If no pattern is provided, list all records of the collection.",

		Args: clapp.Args{
			{
				Name:     "name",
				Usage:    "Name of the record to get information about.",
				Var:      &recordIDs,
				Optional: true,
			},
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if len(recordIDs) == 0 {
				if err := gs.ListAll(); err != nil {
					return err
				}
				return nil
			}

			if err := gs.List(recordIDs); err != nil {
				return err
			}
			return nil
		},
	})

	var query string
	cmd.SubCommands.Add(&clapp.Command{
		Name: "search",

		Usage: "Search the collection's records matching the given query.",

		Args: clapp.Args{
			{
				Name:  "query",
				Usage: "Query to match records against. Query pattern follows blevesearch query language (https://blevesearch.com/docs/Query-String-Query/).",
				Var:   &query,
			},
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.Search(query); err != nil {
				return err
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
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.Edit(recordID); err != nil {
				return err
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
				Var:   &recordIDs,
			},
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.Delete(recordIDs); err != nil {
				return err
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
				Var:   &recordIDs,
			},
			{
				Name:     "dst",
				Usage:    "Destination folder where the record needs to be exported to. Default to current working directory.",
				Var:      &dstFolder,
				Optional: true,
			},
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.Export(recordIDs, dstFolder); err != nil {
				return err
			}
			return nil
		},
	})

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "check",
		Usage: "Verify collection's consistency and repairs or reports found inconsistencies.",

		Flags: clapp.Flags{
			{
				Name:  "delete-ghosts",
				Usage: "Delete any database entries that does not correspond to an existing file in the store's filesystem (so called ghost record)",
				Var:   &cfg.DeleteGhosts,
			},
			{
				Name:  "delete-orphans",
				Usage: "Delete any file of the store's filesystem that is not recorded in the store's database.",
				Var:   &cfg.DeleteOrphans,
			},
			{
				Name:  "import-orphans",
				Usage: "Delete any database entry that does not correspond to an existing file in the store's filesystem (so called ghost record)",
				Var:   &cfg.ImportOrphans,
			},
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.CheckAndRepair(); err != nil {
				return err
			}
			return nil
		},
	})

	return cmd
}
