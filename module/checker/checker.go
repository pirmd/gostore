package checker

import (
	"fmt"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "checker"
)

var (
	_ module.Module = (*checker)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
	// FieldName is the name of the metadata's field to store the quality
	// assessment's result.
	// Default to: QAFindings
	FieldName string

	// MaxFindings is the maximum allowed number of QA findings before the quality check
	// fails.
	// Default to: 0 (disabled)
	MaxFindings int
}

func newConfig() module.Factory {
	return &config{
		FieldName: "QAFindings",
	}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newChecker(cfg, env)
}

// checker is a gostore's module to identify possible quality issue.
type checker struct {
	*module.Environment

	field string
	max   int
}

func newChecker(cfg *config, env *module.Environment) (*checker, error) {
	return &checker{
		Environment: env,
		field:       cfg.FieldName,
		max:         cfg.MaxFindings,
	}, nil
}

// Process completes a record's metadata with a quality level.
func (c *checker) Process(r *store.Record) error {
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

func (c *checker) check(r *store.Record) (map[string]string, error) {
	if r.File() != nil {
		return media.Check(r.Data(), r.File())
	}

	f, err := c.Store.OpenRecord(r)
	if err != nil {
		return nil, err
	}
	f.Close()

	return media.Check(r.Data(), f)
}

func init() {
	module.Register(moduleName, newConfig)
}
