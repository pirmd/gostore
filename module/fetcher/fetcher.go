package fetcher

import (
	"fmt"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	// moduleName of the gostore's module
	moduleName = "fetcher"
)

var (
	_ module.Module = (*fetcher)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
}

func newConfig() module.Factory {
	return &config{}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newFetcher(cfg, env)
}

// fetcher is a gostore's module that retrieves metadata from online databases.
type fetcher struct {
	*module.Environment
}

func newFetcher(cfg *config, env *module.Environment) (*fetcher, error) {
	return &fetcher{
		Environment: env,
	}, nil
}

// Process updates a record's metadata based on the first result returned by
// media.FetchMetadata.
func (f *fetcher) Process(r *store.Record) error {
	matches, err := media.FetchMetadata(r.Data())
	if err != nil {
		return fmt.Errorf("module '%s': fail to fetch metadata: %v", moduleName, err)
	}

	if len(matches) == 0 {
		f.Logger.Printf("Module '%s': no match found, aborting", moduleName)
		return nil
	}

	bestMatch := r.Data()
	for k, v := range matches[0] {
		bestMatch[k] = v
	}

	f.Logger.Printf("Module '%s': found %d match(es), use the first one: %v", moduleName, len(matches), bestMatch)
	r.MergeData(bestMatch)

	return nil
}

func init() {
	module.Register(moduleName, newConfig)
}
