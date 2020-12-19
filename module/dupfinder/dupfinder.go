package dupfinder

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "dupfinder"
)

var (
	_ module.Module = (*dupfinder)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
	// DupQueries is a collection of queries templates that identify records'
	// duplicates.
	// Queries, once expanded, should follow Store.SearchQuery query syntax.
	DupQueries []string
}

func newConfig() module.Factory {
	return &config{}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newDupfinder(cfg, env)
}

// dupfinder is a gostore's module that checks whether a record is already in
// the collection or not.
type dupfinder struct {
	*module.Environment

	finders *template.Template
}

func newDupfinder(cfg *config, env *module.Environment) (*dupfinder, error) {
	d := &dupfinder{
		Environment: env,
		finders:     template.New(moduleName),
	}
	d.finders.Funcs(funcmap)

	for _, txt := range cfg.DupQueries {
		if _, err := d.finders.New("").Parse(txt); err != nil {
			return nil, err
		}
	}

	return d, nil
}

// Process searches for duplicates of a record and fail if any is found.
func (d *dupfinder) Process(r *store.Record) error {
	for _, tmpl := range d.finders.Templates() {
		query := new(bytes.Buffer)
		if err := tmpl.Execute(query, r.Flatted()); err != nil {
			return fmt.Errorf("module '%s': fail to look for duplicate: %v", moduleName, err)
		}

		matches, err := d.Store.SearchQuery(query.String())
		if err != nil {
			return fmt.Errorf("module '%s': fail to look for duplicate: %v", moduleName, err)
		}
		if len(matches) > 0 {
			return fmt.Errorf("module '%s': possible duplicate(s) of record (%v) found in the database", moduleName, matches)
		}
	}

	return nil
}

func init() {
	module.Register(moduleName, newConfig)
}
