package modules

import (
	"errors"
	"log"

	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui"
)

var (
	// ErrUnknownModule error is raised when the supplied module name is not
	// found in the module register.
	ErrUnknownModule = errors.New("unknown module")

	// availableModules contains all available gostore's modules.
	availableModules factories
)

// Module represents a gostore's module that can act on collection's records.
type Module interface {
	// ProcessRecord performs an action on a collection's record
	ProcessRecord(*store.Record) error
}

// Unmarshaler represents an encoded type that know how to unmarshal itself.
// Module package relies on this interface for unmarshaling raw configuration
// into a given module's config.
type Unmarshaler interface {
	// Unmarshal decodes a raw configuration into a module's config
	Unmarshal(interface{}) error
}

// Environment represents a set of facilities that a module can access to to
// interact with its environment (logging, user interaction, store access).
type Environment struct {
	// Logger provides logging facility to a module.
	Logger *log.Logger
	// UI provides user interaction facilities to a module.
	UI ui.UserInterfacer
	// Store provides a collection's access facilities to a module.
	Store *store.Store
}

// factory represents a module provider.
type factory struct {
	// Name of the module.
	Name string
	// NewModule creates a new instance of the module.
	NewModule func(rawcfg Unmarshaler, env *Environment) (Module, error)
}

type factories []*factory

func (f *factories) append(fact *factory) {
	*f = append(*f, fact)
}

func (f *factories) get(name string) *factory {
	for _, factory := range *f {
		if factory.Name == name {
			return factory
		}
	}
	return nil
}

// New returns a new module instance of module "name".
// If the provided name is not a registered module, ErrUnknownModule is raised.
func New(name string, rawcfg Unmarshaler, env *Environment) (Module, error) {
	factory := availableModules.get(name)
	if factory == nil {
		return nil, ErrUnknownModule
	}

	return factory.NewModule(rawcfg, env)
}

// Register registers a new Module.
// If a Module already exists for the supplied name, it will be replaced by the
// new provided module's factory.
func Register(name string, newFn func(Unmarshaler, *Environment) (Module, error)) {
	availableModules.append(&factory{
		Name:      name,
		NewModule: newFn,
	})
}

// List returns all available modules names.
func List() (m []string) {
	for _, factory := range availableModules {
		m = append(m, factory.Name)
	}
	return
}

// ProcessRecord applies the specified series of modules to a record. If an
// error occurs, ProcessRecord stops and feedback the error.
func ProcessRecord(r *store.Record, mods []Module) error {
	for _, mod := range mods {
		if err := mod.ProcessRecord(r); err != nil {
			return err
		}
	}

	return nil
}
