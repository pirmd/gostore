// Package mdatareader reads metadata from a media file and populates the
// corresponding Record's values.
package mdatareader

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui"
)

const (
	moduleName = "mdatareader"
)

var (
	_ modules.Module = (*mdataReader)(nil) // Makes sure that we implement modules.Module interface.
)

// Config defines the different module's options.
type Config struct {
}

func newConfig() *Config {
	return &Config{}
}

type mdataReader struct {
	log *log.Logger
	ui  ui.UserInterfacer
}

func newMdataReader(cfg *Config, logger *log.Logger, UI ui.UserInterfacer) (modules.Module, error) {
	return &mdataReader{
		log: logger,
		ui:  UI,
	}, nil
}

// ProcessRecord completes a record's metadata with a quality level.
func (m *mdataReader) ProcessRecord(r *store.Record) error {
	if r.File() == nil {
		m.log.Printf("Module '%s': no record's file available for %s", moduleName, r.Key())
		return nil
	}

	mdataFromFile, err := media.ReadMetadata(r.File())
	if err != nil {
		return fmt.Errorf("module '%s': fail to read metadata for '%s': %v", moduleName, r.Key(), err)
	}

	m.log.Printf("Module '%s': found metadata: %v", moduleName, mdataFromFile)
	mdata, err := m.ui.Merge(mdataFromFile, r.Data())
	if err != nil {
		return fmt.Errorf("module '%s': fail to merge metadata: %v", moduleName, err)
	}

	m.log.Printf("Module '%s': record updated to: %v", moduleName, mdata)
	r.SetData(mdata)

	return nil
}

// NewFromRawConfig creates a new module from a raw configuration.
func NewFromRawConfig(rawcfg modules.Unmarshaler, env *modules.Environment) (modules.Module, error) {
	env.Logger.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newMdataReader(cfg, env.Logger, env.UI)
}

func init() {
	modules.Register(moduleName, NewFromRawConfig)
}
