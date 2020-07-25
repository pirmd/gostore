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
	// FieldName is the name of the metadata field where to store checker
	// outcome. Default to: QALevel.
	FieldName string
	// MinLevel is the minimum allowed level. Any level below this threshold
	// will result in an error aborting operation. Default to 0 (all quality
	// level is accepted).
	MinLevel int
}

func newConfig() *Config {
	return &Config{
		FieldName: "QALevel",
	}
}

type checker struct {
	log *log.Logger

	field    string
	minLevel int
}

func newChecker(cfg *Config, logger *log.Logger) (modules.Module, error) {
	return &checker{
		log:      logger,
		field:    cfg.FieldName,
		minLevel: cfg.MinLevel,
	}, nil
}

// ProcessRecord completes a record's metadata with a quality level.
func (c *checker) ProcessRecord(r *store.Record) error {
	c.log.Printf("Module '%s': assess quality level", moduleName)
	lvl := media.CheckMetadata(r.Data())
	r.Set(c.field, lvl)

	if lvl < c.minLevel {
		return fmt.Errorf("module '%s': minimum level of quality is not reached (got %d, min %d)", moduleName, lvl, c.minLevel)
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

	return newChecker(cfg, env.Logger)
}

func init() {
	modules.Register(moduleName, NewFromRawConfig)
}
