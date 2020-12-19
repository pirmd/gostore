package scrubber

import (
	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "scrubber"
)

var (
	_ module.Module = (*scrubber)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
	// Fields lists the record's fields that should be scrubbed.
	Fields []string
}

func newConfig() module.Factory {
	return &config{}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newScrubber(cfg, env)
}

// scrubber is a gostore's module that removes fields from a set of metadata.
type scrubber struct {
	*module.Environment

	fields []string
}

func newScrubber(cfg *config, env *module.Environment) (*scrubber, error) {
	return &scrubber{
		Environment: env,
		fields:      cfg.Fields,
	}, nil
}

// Process scrubs records from unneeded/unwanted fields.
func (s *scrubber) Process(r *store.Record) error {
	for _, field := range s.fields {
		s.Logger.Printf("Module '%s': scrub field '%s'", moduleName, field)
		r.Del(field)
	}
	return nil
}

func init() {
	module.Register(moduleName, newConfig)
}
