//go:generate go run manpage_generate.go cmd.go gostore.go config.go
package main

import (
	"fmt"
	"strings"

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
			Name:  "verbose",
			Usage: "Show verbose information.",
			Var:   &cfg.Verbose,
		},

		{
			Name:  "debug",
			Usage: "Show debug information.",
			Var:   &cfg.Debug,
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

	var recordIDs []string
	var recordIDsArg = &clapp.Arg{
		Name:     "name",
		Usage:    "Record's name. Name can be specified using a glob pattern.",
		Var:      &recordIDs,
		Optional: true,
	}

	var sortBy []string
	var sortByFlag = &clapp.Flag{
		Name:  "sort",
		Usage: "Sort the search results. Record will first be sorted by the first field. Any items with the same value for that field, are then also sorted by the next field, and so on. The names of fields can be prefixed with the - character, which will cause that field to be reversed (descending order). Special fields are provided '_id' (record's name) and '_score' (search relevance score).",
		Var:   &sortBy,
	}

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "list",
		Usage: "List and retrieve information about collection's records. If no pattern is provided, list all records of the collection.",

		Args: clapp.Args{
			recordIDsArg,
		},

		Flags: clapp.Flags{
			sortByFlag,
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if len(recordIDs) == 0 {
				if err := gs.ListAll(sortBy); err != nil {
					return err
				}
				return nil
			}

			if err := gs.ListGlob(recordIDs, sortBy); err != nil {
				return err
			}
			return nil
		},
	})

	var query string
	cmd.SubCommands.Add(&clapp.Command{
		Name:  "search",
		Usage: "Search the collection's records matching the given query.",

		Args: clapp.Args{
			{
				Name:  "query",
				Usage: "Query to match records against. Query pattern follows blevesearch query language (https://blevesearch.com/docs/Query-String-Query/).",
				Var:   &query,
			},
		},

		Flags: clapp.Flags{
			sortByFlag,
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.ListQuery(query, sortBy); err != nil {
				return err
			}
			return nil
		},
	})

	var multiEdit bool
	cmd.SubCommands.Add(&clapp.Command{
		Name:  "edit",
		Usage: "Edit an existing record from the collection using user defined's editor. If flag '--auto' is used, edition is skipped and nothing happens.",

		Flags: clapp.Flags{
			{
				Name:  "multi-edit",
				Usage: "Edit multiple records at once instead of individually. Make sure when editing to not modify records order not do delete or add one.",
				Var:   &multiEdit,
			},
			{
				Name:  "import-orphans",
				Usage: "Delete any database entry that does not correspond to an existing file in the store's filesystem (so called ghost record)",
				Var:   &cfg.ImportOrphans,
			},
		},

		Args: clapp.Args{
			recordIDsArg,
		},

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if multiEdit {
				if err := gs.MultiEdit(recordIDs); err != nil {
					return err
				}
				return nil
			}

			if err := gs.Edit(recordIDs); err != nil {
				return err
			}
			return nil
		},
	})

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "delete",
		Usage: "Delete an existing record from the collection.",

		Args: clapp.Args{
			recordIDsArg,
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
			recordIDsArg,
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

			if err := gs.Export(dstFolder, recordIDs); err != nil {
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

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "rebuild-index",
		Usage: "Deletes then rebuild the collection's index from scratch. Useful for example to implement a new mapping strategy or if things are really going bad.",

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.RebuildIndex(); err != nil {
				return err
			}
			return nil
		},
	})

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "fields",
		Usage: "Lists fields names that are available for search or for templates. Some fields might only be available for a given media Type.",

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			if err := gs.Fields(); err != nil {
				return err
			}
			return nil
		},
	})

	cmd.SubCommands.Add(&clapp.Command{
		Name:  "config",
		Usage: "Lists known modules, analyzers or styles that can be used for gostore configuration or invocation.",

		Execute: func() error {
			gs, err := openGostore(cfg)
			if err != nil {
				return err
			}
			defer gs.Close()

			fmt.Printf("Known Modules:\n%s\n", strings.Join(cfg.Modules(), "\n"))
			fmt.Printf("\nKnown Analyzers:\n%s\n", strings.Join(cfg.Analyzers(), "\n"))
			fmt.Printf("\nKnown Styles:\n%s\n", strings.Join(cfg.Styles(), "\n"))
			return nil
		},
	})

	return cmd
}
