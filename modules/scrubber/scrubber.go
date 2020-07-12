// Package scrubber removes fields from a set of metadata.
package scrubber

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui"
)

const (
	moduleName = "scrubber"
)

var (
	_ modules.Module = (*scrubber)(nil) // Makes sure that scrubber implements modules.Module
)

// Config defines the different configurations that can be used to customized
// the behavior of a scrubber module.
type Config struct {
	// Fields lists the record's fields that should be scrubbed.
	Fields []string
}

func newConfig() *Config {
	return &Config{}
}

type scrubber struct {
	log *log.Logger

	fields []string
}

func newScrubber(cfg *Config, logger *log.Logger) (modules.Module, error) {
	return &scrubber{
		log:    logger,
		fields: cfg.Fields,
	}, nil
}

// ProcessRecord scrubs records from unneeded/unwanted fields.
func (s *scrubber) ProcessRecord(r *store.Record) error {
	for _, field := range s.fields {
		s.log.Printf("Module '%s': scrub field '%s'", moduleName, field)
		r.Del(field)
	}
	return nil
}

// New creates a new scrubber module whose configuration information
func New(rawcfg modules.ConfigUnmarshaler, logger *log.Logger, UI ui.UserInterfacer) (modules.Module, error) {
	logger.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newScrubber(cfg, logger)
}

func init() {
	modules.Register(moduleName, New)
}
