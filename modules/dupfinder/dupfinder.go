// Package dupfinder is a module that checks whether a record is already in the
// collection or not.
package dupfinder

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "dupfinder"
)

var (
	_ modules.Module = (*dupfinder)(nil) // Makes sure that we implement modules.Module interface.
)

// Config defines the different module's options.
type Config struct {
	// DupQueries is a collection of queries templates that identify records'
	// duplicates.
	// Queries, once expanded, should follow Store.SearchQuery query syntax.
	DupQueries []string
}

func newConfig() *Config {
	return &Config{}
}

type dupfinder struct {
	log     *log.Logger
	store   *store.Store
	finders *template.Template
}

func newDupfinder(cfg *Config, logger *log.Logger, store *store.Store) (modules.Module, error) {
	d := &dupfinder{
		log:     logger,
		store:   store,
		finders: template.New("dupfinder"),
	}
	d.finders.Funcs(funcmap)

	for _, txt := range cfg.DupQueries {
		if _, err := d.finders.New("").Parse(txt); err != nil {
			return nil, err
		}
	}

	return d, nil
}

// ProcessRecord searches for duplicates of a records and failed if any is found.
func (d *dupfinder) ProcessRecord(r *store.Record) error {
	for _, tmpl := range d.finders.Templates() {
		query := new(bytes.Buffer)
		if err := tmpl.Execute(query, r.Flatted()); err != nil {
			return fmt.Errorf("module '%s': fail to look for duplicate: %v", moduleName, err)
		}

		matches, err := d.store.SearchQuery(query.String())
		if err != nil {
			return fmt.Errorf("module '%s': fail to look for duplicate: %v", moduleName, err)
		}
		if len(matches) > 0 {
			return fmt.Errorf("module '%s': possible duplicate(s) of record (%v) found in the database", moduleName, matches)
		}
	}

	return nil
}

// NewFromRawConfig creates a new module from a raw configuration.
func NewFromRawConfig(rawcfg modules.Unmarshaler, env *modules.Environment) (modules.Module, error) {
	env.Logger.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newDupfinder(cfg, env.Logger, env.Store)
}

func init() {
	modules.Register(moduleName, NewFromRawConfig)
}
