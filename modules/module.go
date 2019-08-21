package modules

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/store"
)

var (
	//ErrUnknownModule error is raised when the supplied module name is not
	//found in the module register.
	ErrUnknownModule = fmt.Errorf("unknown module")

	//modulesRegister contains all available gostore's modules
	modulesRegister = make(map[string]NewModuleFn)
)

//NewModuleFn specifies the standard way to initiate a new module
type NewModuleFn func(cfg interface{}, logger *log.Logger) (Module, error)

//Module represents a gostore's module that can act on collection's records
type Module interface {
	//ProcessRecord performs an action on a collection's record
	ProcessRecord(*store.Record) error
}

//New returns a new module instance. If the provided nam is not registered,
//ErrUnknownModule is raised.
func New(name string, cfg interface{}, logger *log.Logger) (Module, error) {
	newFn, exists := modulesRegister[name]
	if exists {
		return nil, ErrUnknownModule
	}

	return newFn(cfg, logger)
}

//Register registers a new Module.
//If a Module already exists for the supplied name, it will be replaced by the
//new provided module
func Register(name string, newFn NewModuleFn) {
	modulesRegister[name] = newFn
}

//ProcessRecord applies the specified series of modules to a record. If an
//error occures, ProcessRecord stops and feedback the error
func ProcessRecord(r *store.Record, mods []Module) error {
	for _, mod := range mods {
		if err := mod.ProcessRecord(r); err != nil {
			return err
		}
	}

	return nil
}
