package module

import (
	"encoding/json"
	"fmt"
)

// Config represents a module configuration.
type Config struct {
	Factory
}

// UnmarshalJSON implements JSON interface to customize JSON unmarshalling
// process. It detects the module name (field: Module) and generates the
// appropriate module's constructor.
func (c *Config) UnmarshalJSON(data []byte) error {
	return c.unmarshal(func(v interface{}) error {
		return json.Unmarshal(data, &v)
	})
}

// UnmarshalYAML implements YAML interface to customize JSON unmarshalling
// process. It detects the module name (field: Module) and generates the
// appropriate module's constructor.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return c.unmarshal(unmarshal)
}

func (c *Config) unmarshal(unmarshal func(interface{}) error) error {
	id := struct {
		Module string
	}{}

	if err := unmarshal(&id); err != nil {
		return err
	}

	newFact, exists := modules[id.Module]
	if !exists {
		return fmt.Errorf("cannot configure '%s': unknown module", id.Module)
	}
	c.Factory = newFact()

	if err := unmarshal(c.Factory); err != nil {
		return err
	}

	return nil
}
