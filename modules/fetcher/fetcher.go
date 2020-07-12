// Package fetcher is a gostore module that retrieves metadata from online
// databases.
package fetcher

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
	moduleName = "fetcher"
)

var (
	_ modules.Module = (*fetcher)(nil) // Makes sure that we implement modules.Module
)

// Config defines the different configurations that can be used to customize
// the behavior of the fetcher module.
type Config struct {
}

func newConfig() *Config {
	return &Config{}
}

type fetcher struct {
	log *log.Logger
	ui  ui.UserInterfacer
}

func newFetcher(cfg *Config, logger *log.Logger, UI ui.UserInterfacer) (*fetcher, error) {
	return &fetcher{
		log: logger,
		ui:  UI,
	}, nil
}

// ProcessRecord updates a record's metadata based on the first result returned
// by media.FetchMetadata.
func (f *fetcher) ProcessRecord(r *store.Record) error {
	f.log.Printf("Module '%s': fetch metadata for '%v'", moduleName, r.Data())
	matches, err := media.FetchMetadata(r.Data())
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		f.log.Printf("Module '%s': no match found, aborting", moduleName)
		return nil
	}

	bestMatch := r.Data()
	for k, v := range matches[0] {
		bestMatch[k] = v
	}

	f.log.Printf("Module '%s': found %d match(es), use the first one: %v", moduleName, len(matches), bestMatch)
	mdata, err := f.ui.Merge(bestMatch, r.Data())
	if err != nil {
		return err
	}

	f.log.Printf("Module '%s': record updated to: %v", moduleName, mdata)
	r.SetData(mdata)

	return nil
}

// New creates a new fetcher module
func New(rawcfg modules.ConfigUnmarshaler, logger *log.Logger, UI ui.UserInterfacer) (modules.Module, error) {
	logger.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newFetcher(cfg, logger, UI)
}

func init() {
	modules.Register(moduleName, New)
}
