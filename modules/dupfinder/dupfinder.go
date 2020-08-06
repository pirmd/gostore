// Package dupfinder is a module that checks whether a record is already in the
// collection.
package dupfinder

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/media"
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
	// SimilarityLevel is the measure of similarity that is accepted
	// between two records. Default to 1.
	SimilarityLevel int
}

func newConfig() *Config {
	return &Config{
		SimilarityLevel: 1,
	}
}

type dupfinder struct {
	log       *log.Logger
	store     *store.Store
	fuzziness int
}

func newDupfinder(cfg *Config, logger *log.Logger, store *store.Store) (modules.Module, error) {
	return &dupfinder{
		log:       logger,
		store:     store,
		fuzziness: cfg.SimilarityLevel,
	}, nil
}

// ProcessRecord completes a record's metadata with a quality level.
func (d *dupfinder) ProcessRecord(r *store.Record) error {
	exact, similar := media.IDCard(r.Data())
	d.log.Printf("Module '%s': find dupplicate(s) for EXACT=%v or SIMILAR=%v (similarity lvl=%d)", moduleName, exact, similar, d.fuzziness)

	matches, err := d.store.SearchFields(exact, 0)
	if err != nil {
		return fmt.Errorf("module '%s': fail to look for dupplicate: %v", moduleName, err)
	}
	if len(matches) > 0 {
		return fmt.Errorf("module '%s': possible dupplicate(s) of record (%v) found in the database", moduleName, matches)
	}

	matches, err = d.store.SearchFields(similar, d.fuzziness)
	if err != nil {
		return fmt.Errorf("module '%s': fail to look for dupplicate: %v", moduleName, err)
	}
	if len(matches) > 0 {
		return fmt.Errorf("module '%s': dupplicate of record (%v) already existing in the database", moduleName, matches)
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
