// Package scrubber removes fields from a set of metadata.
package scrubber

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "scrubber"
)

var (
	_ modules.Module = (*scrubber)(nil) // Makes sure that we implement modules.Module interface.
)

// Config defines the different module's options.
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

// NewFromRawConfig creates a new module from a raw configuration.
func NewFromRawConfig(rawcfg modules.Unmarshaler, env *modules.Environment) (modules.Module, error) {
	env.Logger.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newScrubber(cfg, env.Logger)
}

func init() {
	modules.Register(moduleName, NewFromRawConfig)
}
