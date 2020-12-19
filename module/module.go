package module

import (
	"log"

	"github.com/pirmd/gostore/store"
)

var (
	// modules contains all available gostore's modules.
	modules = make(map[string]func() Factory)
)

// Module represents a gostore's module that can act on collection's records.
type Module interface {
	// Process performs an action on a collection's record
	Process(*store.Record) error
}

// Factory represents a module provider.
type Factory interface {
	// NewModule creates a new module according to Factory configuration and to
	// gostore's modules environment.
	NewModule(*Environment) (Module, error)
}

// Environment represents a set of facilities that a module can access to to
// interact with its environment (logging, gostore access, ...).
type Environment struct {
	// Logger provides logging facility to a module.
	Logger *log.Logger
	// Store provides a collection's access facilities to a module.
	Store *store.Store
}

// Register registers a new Module Factory.
// If a Module Factory already exists for the supplied name, it will be
// replaced.
func Register(name string, newFactory func() Factory) {
	modules[name] = newFactory
}

// New creates a new Module.
func New(cfg *Config, env *Environment) (Module, error) {
	return cfg.NewModule(env)
}

// List returns all available modules names.
func List() (m []string) {
	for mod := range modules {
		m = append(m, mod)
	}
	return
}
