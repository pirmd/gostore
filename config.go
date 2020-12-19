package main

import (
	"os"

	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui/cli"
)

// Config represents the configuration for gostore.
type Config struct {
	// Verbose is a flag that governs if log information are to be shown
	Verbose bool

	// Debug is a flag that governs if debug information are to be shown
	Debug bool

	// ReadOnly is the flag to switch the store into read only operation mode
	ReadOnly bool

	// DeleteGhosts is a flag that instructs gostore.Check to delete any
	// database entries that does not correspond to an existing file in the
	// store's filesystem (so called ghost record)
	DeleteGhosts bool

	// DeleteOrphans is a flag that instructs gostore.Check to delete any
	// file of the store's filesystem that is not recorded in the store's
	// database.
	DeleteOrphans bool

	// ImportOrphans is a flag that instructs gostore.Check to re-import any
	// file of the store's filesystem that is not recorded in the store's
	// database.
	ImportOrphans bool

	// Store contains configuration for anything related to storage
	Store *store.Config

	// UI contains configuration for anything related to user interface
	UI *cli.Config

	// Import list of actions to apply when importing a record
	Import []*module.Config

	// Update list of actions to apply when importing a record
	Update []*module.Config
}

// Modules lists available module.
func (cfg *Config) Modules() []string {
	return module.List()
}

// Analyzers lists available analyzers for indexing records.
func (cfg *Config) Analyzers() []string {
	return cfg.Store.Analyzers()
}

// Styles lists available styles for printing records.
func (cfg *Config) Styles() []string {
	return cfg.UI.Styles()
}

func (cfg *Config) expandEnv() {
	cfg.Store.Path = os.ExpandEnv(cfg.Store.Path)
	cfg.UI.EditorCmd = os.ExpandEnv(cfg.UI.EditorCmd)
	cfg.UI.MergerCmd = os.ExpandEnv(cfg.UI.MergerCmd)
}

func newConfig() *Config {
	return &Config{
		Store: store.NewConfig(),
		UI:    cli.NewConfig(),
	}
}
