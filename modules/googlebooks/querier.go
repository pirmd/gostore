// Package googlebooks is a gostore module that retrieves ebook metadata from
// google books online database.
package googlebooks

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	// moduleName of the gostore's module
	moduleName = "GoogleBooksQuerier"
)

var (
	// Makes sure that we implement modules.Module
	_ modules.Module = (*querier)(nil)
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
}

func newQuerier(cfg *Config, logger *log.Logger) (*querier, error) {
	return &querier{
		log:         logger,
		googleBooks: &googleBooks{},
	}, nil
}

// ProcessRecord updates a record's metadata based on the first result returned
// by google books.
// If provided record is not an ebook, record metadata is not modified.
// If no result is found, record metadata is not modified.
func (q *querier) ProcessRecord(r *store.Record) error {
	if !media.IsOfType(r.Value(), "book") {
		q.log.Printf("Module '%s': record is not an ebook, aborting", moduleName)
		return nil
	}

	q.log.Printf("Module '%s': query GoogleBooks for '%v'", moduleName, r.Value())
	found, err := q.LookForBooks(r.Value())
	if err != nil {
		return err
	}

	if len(found) == 0 {
		q.log.Printf("Module '%s': no match found, aborting", moduleName)
		return nil
	}

	//TODO(pirmd): do something more clever than using the first result from
	//googleBooks
	q.log.Printf("Module '%s': found %d match(es), use the first one: %v", moduleName, len(found), found[0])
	r.MergeValues(found[0])
	return nil
}

// New creates a new GoogleBooksQuerier module
func New(rawcfg modules.ConfigUnmarshaler, log *log.Logger) (modules.Module, error) {
	log.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newQuerier(cfg, log)
}

func init() {
	modules.Register(moduleName, New)
}
