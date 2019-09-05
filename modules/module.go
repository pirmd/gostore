package modules

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pirmd/gostore/store"
)

var (
	//ConfUnmarshal is the default module configuration parser
	ConfUnmarshal ConfUnmarshaller = json.Unmarshal

	//ErrUnknownModule error is raised when the supplied module name is not
	//found in the module register.
	ErrUnknownModule = fmt.Errorf("unknown module")

	//modulesRegister contains all available gostore's modules
	modulesRegister = make(map[string]NewModuleFn)
)

//NewModuleFn specifies the standard way to initiate a new module
type NewModuleFn func(conf []byte, logger *log.Logger) (Module, error)

//ConfUnmarshaller parses a text-based module's configuration into a module
//configuration struct
type ConfUnmarshaller func(conf []byte, cfg interface{}) error

//Module represents a gostore's module that can act on collection's records
type Module interface {
	//ProcessRecord performs an action on a collection's record
	ProcessRecord(*store.Record) error
}

//New returns a new module instance of module "name" with configuration given
//in a text-based format.
//Module configuration format should match ConfUnmarshal encoding idiom (JSON,
//YAML, TOML...).
//If the provided name is not a registered module, ErrUnknownModule is raised.
func New(name string, conf []byte, logger *log.Logger) (Module, error) {
	newModule, exists := modulesRegister[name]
	if exists {
		return nil, ErrUnknownModule
	}

	return newModule(conf, logger)
}

//Register registers a new Module.
//If a Module already exists for the supplied name, it will be replaced by the
//new provided module.
func Register(name string, newFn NewModuleFn) {
	modulesRegister[name] = newFn
}

//ProcessRecord applies the specified series of modules to a record. If an
//error occures, ProcessRecord stops and feedback the error.
func ProcessRecord(r *store.Record, mods []Module) error {
	for _, mod := range mods {
		if err := mod.ProcessRecord(r); err != nil {
			return err
		}
	}

	return nil
}
