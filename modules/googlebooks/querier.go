// Package googlebooks is a gostore module that retrieves ebook metadata from
// Google books online database.
package googlebooks

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui"
)

const (
	// moduleName of the gostore's module
	moduleName = "GoogleBooksQuerier"
)

var (
	_ modules.Module = (*querier)(nil) // Makes sure that we implement modules.Module
)

// Config defines the different configurations that can be used to customize
// the behavior of an organizer module.
type Config struct {
}

func newConfig() *Config {
	return &Config{}
}

type querier struct {
	*googleBooks

	log *log.Logger
	ui  ui.UserInterfacer
}

func newQuerier(cfg *Config, logger *log.Logger, UI ui.UserInterfacer) (*querier, error) {
	return &querier{
		log:         logger,
		ui:          UI,
		googleBooks: &googleBooks{},
	}, nil
}

// ProcessRecord updates a record's metadata based on the first result returned
// by Google books.
// If provided record is not an ebook, record metadata is not modified.
// If no result is found, record metadata is not modified.
func (q *querier) ProcessRecord(r *store.Record) error {
	if !media.IsOfType(r.Value(), "book") {
		q.log.Printf("Module '%s': record is not an ebook, aborting", moduleName)
		return nil
	}

	q.log.Printf("Module '%s': query GoogleBooks for '%v'", moduleName, r.Value())
	matches, err := q.LookForBooks(r.Value())
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		q.log.Printf("Module '%s': no match found, aborting", moduleName)
		return nil
	}

	//TODO(pirmd): do something cleverer than using the first result from
	//googleBooks
	bestMatch := r.UserValue()
	for k, v := range matches[0] {
		bestMatch[k] = v
	}

	q.log.Printf("Module '%s': found %d match(es), use the first one: %v", moduleName, len(matches), bestMatch)
	mdata, err := q.ui.Merge(bestMatch, r.UserValue())
	if err != nil {
		return err
	}

	q.log.Printf("Module '%s': record updated to: %v", moduleName, mdata)
	r.SetValue(mdata)

	return nil
}

// New creates a new GoogleBooksQuerier module
func New(rawcfg modules.ConfigUnmarshaler, log *log.Logger, UI ui.UserInterfacer) (modules.Module, error) {
	log.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newQuerier(cfg, log, UI)
}

func init() {
	modules.Register(moduleName, New)
}
