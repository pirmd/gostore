package mdatareader

import (
	"fmt"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "mdatareader"
)

var (
	_ module.Module = (*mdataReader)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
}

func newConfig() module.Factory {
	return &config{}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newMdataReader(cfg, env)
}

// mdataReader is a gostore's module that reads metadata from a media file and
// populates the corresponding Record's values.
type mdataReader struct {
	*module.Environment
}

func newMdataReader(cfg *config, env *module.Environment) (*mdataReader, error) {
	return &mdataReader{
		Environment: env,
	}, nil
}

// Process completes a record's metadata with a quality level.
func (m *mdataReader) Process(r *store.Record) error {
	if r.File() == nil {
		m.Logger.Printf("Module '%s': no record's file available for %s", moduleName, r.Key())
		return nil
	}

	mdataFromFile, err := media.ReadMetadata(r.File())
	if err != nil {
		return fmt.Errorf("module '%s': fail to read metadata for '%s': %v", moduleName, r.Key(), err)
	}

	m.Logger.Printf("Module '%s': found metadata: %v", moduleName, mdataFromFile)
	r.MergeData(mdataFromFile)

	return nil
}

func init() {
	module.Register(moduleName, newConfig)
}
