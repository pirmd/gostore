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
	// Name of the module
	Name = "GoogleBooksQuerier"
)

var (
	// Makes sure that we implement modules.Module
	_ modules.Module = (*querier)(nil)

	// ErrNoMetadataFound reports an error when no Metadata found
	ErrNoMetadataFound = fmt.Errorf("module '%s': no metadata found", Name)
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

// ProcessRecord modifies the record's name to match a standardized naming scheme.
func (q *querier) ProcessRecord(r *store.Record) error {
	typ, ok := r.GetValue(media.TypeField).(string)
	if !ok {
		return fmt.Errorf("module '%s': invalid or unknown record's type", Name)
	}
	//TODO(pirmd): make it better way to capture it, or media/books.Types()
	//             maybe have googlebooks depends of media/books and
	//             provide feedback in case of unaccurate media type
	if typ != "epub" {
		return nil
		//XXX: return ErrNotSupportedMEdiaType
	}

	found, err := q.LookForBooks(r.Value())
	if err != nil {
		return err
	}

	if len(found) == 0 {
		return ErrNoMetadataFound
	}

	//TODO(pirmd): do something clever than using the first result from
	//googleBooks
	r.MergeValues(found[0])
	return nil
}

// New creates a new organizer module
func New(rawcfg modules.ConfigUnmarshaler, log *log.Logger) (modules.Module, error) {
	log.Printf("Module '%s': new logger with config '%v'", Name, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", Name, err)
	}

	return newQuerier(cfg, log)
}

func init() {
	modules.Register(Name, New)
}
