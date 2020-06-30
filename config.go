package main

import (
	"os"

	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui/cli"
)

// Config represents the configuration for gostore
type Config struct {
	// ShowLog is a flag that governs if log information are to be shown
	ShowLog bool

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
	Import []*moduleConfig

	// Update list of actions to apply when importing a record
	Update []*moduleConfig
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

type moduleConfig struct {
	Name   string
	Config *rawYAMLConfig
}

type rawYAMLConfig struct {
	unmarshal func(interface{}) error
}

func (cfg *rawYAMLConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	cfg.unmarshal = unmarshal
	return nil
}

func (cfg *rawYAMLConfig) Unmarshal(v interface{}) error {
	if cfg == nil {
		return nil
	}

	return cfg.unmarshal(v)
}
