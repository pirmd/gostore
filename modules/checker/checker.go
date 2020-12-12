// Package checker assesses the quality level of a record on a 0 to 100 scale
// (0: very bad, 100: perfect).
package checker

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "checker"
)

var (
	_ modules.Module = (*checker)(nil) // Makes sure that we implement modules.Module interface.
)

// Config defines the different module's options.
type Config struct {
	// FieldName is the name of the metadata's field to store the quality
	// assessment's result.
	// Default to: QAFindings
	FieldName string

	// MaxFindings is the maximum allowed number of QA findings before the quality check
	// fails.
	// Default to: 0 (disabled)
	MaxFindings int
}

func newConfig() *Config {
	return &Config{
		FieldName: "QAFindings",
	}
}

type checker struct {
	log   *log.Logger
	store *store.Store

	field string
	max   int
}

func newChecker(cfg *Config, logger *log.Logger, storer *store.Store) (modules.Module, error) {
	return &checker{
		log:   logger,
		store: storer,
		field: cfg.FieldName,
		max:   cfg.MaxFindings,
	}, nil
}

// ProcessRecord completes a record's metadata with a quality level.
func (c *checker) ProcessRecord(r *store.Record) error {
	c.log.Printf("Module '%s': assess quality level", moduleName)

	issues, err := c.check(r)
	if err != nil {
		return fmt.Errorf("module '%s': fail to assess quality level: %v", moduleName, err)
	}

	if c.max > 0 && len(issues) > c.max {
		return fmt.Errorf("module '%s': minimum level of quality is not reached (got %d, min %d): %s", moduleName, len(issues), c.max, issues)
	}

	if len(issues) > 0 {
		r.Set(c.field, issues)
	} else {
		r.Del(c.field)
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

	return newChecker(cfg, env.Logger, env.Store)
}

func (c *checker) check(r *store.Record) (map[string]string, error) {
	if r.File() != nil {
		return media.Check(r.Data(), r.File())
	}

	f, err := c.store.OpenRecord(r)
	if err != nil {
		return nil, err
	}
	f.Close()

	return media.Check(r.Data(), f)
}

func init() {
	modules.Register(moduleName, NewFromRawConfig)
}
