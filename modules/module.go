package modules

import (
	"errors"
	"log"

	"github.com/pirmd/gostore/store"
)

var (
	// ErrUnknownModule error is raised when the supplied module name is not
	// found in the module register.
	ErrUnknownModule = errors.New("unknown module")

	// availableModules contains all available gostore's modules
	availableModules factories
)

// Module represents a gostore's module that can act on collection's records
type Module interface {
	//ProcessRecord performs an action on a collection's record
	ProcessRecord(*store.Record) error
}

// ConfigUnmarshaler represents a raw configuration that can be unmarshalled to
// a given module's config.
type ConfigUnmarshaler interface {
	// Unmarshal decodes a raw configuration into a module's config
	Unmarshal(interface{}) error
}

// factory represents a module provider.
type factory struct {
	// Name of the module
	Name string

	// NewModule creates a new instance of the module
	NewModule func(rawcfg ConfigUnmarshaler, logger *log.Logger) (Module, error)
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
func New(name string, rawcfg ConfigUnmarshaler, logger *log.Logger) (Module, error) {
	factory := availableModules.get(name)
	if factory == nil {
		return nil, ErrUnknownModule
	}

	return factory.NewModule(rawcfg, logger)
}

// Register registers a new Module.
// If a Module already exists for the supplied name, it will be replaced by the
// new provided module's factory.
func Register(name string, newFn func(ConfigUnmarshaler, *log.Logger) (Module, error)) {
	availableModules.append(&factory{
		Name:      name,
		NewModule: newFn,
	})
}

// List returns all availbale modules names
func List() (m []string) {
	for _, factory := range availableModules {
		m = append(m, factory.Name)
	}
	return
}

// ProcessRecord applies the specified series of modules to a record. If an
// error occures, ProcessRecord stops and feedback the error.
func ProcessRecord(r *store.Record, mods []Module) error {
	for _, mod := range mods {
		if err := mod.ProcessRecord(r); err != nil {
			return err
		}
	}

	return nil
}
