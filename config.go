package main

// Config represents the configuration for gostore
type Config struct {
	//ShowLog is a flag that governs if log information are to be shown
	ShowLog bool
	//Store contains configuration for anything related to storage
	Store *storeConfig
	//UI contains configuration for anything related to user interface
	UI *CLIConfig
	//ImportModules lists of processings to be applied when importing a record
	ImportModules map[string]*rawYAMLConfig
	//UpdateModules lists of processings to be applied when updating a record
	UpdateModules map[string]*rawYAMLConfig
}

// storeConfig contains configuration for storage
type storeConfig struct {
	//Root contains the path to the datastore
	Root string
	//ReadOnly is the flag to switch the store into read only operation mode
	ReadOnly bool
}

func newConfig() *Config {
	return &Config{
		Store:         &storeConfig{Root: "."},
		UI:            &CLIConfig{},
		ImportModules: make(map[string]*rawYAMLConfig),
		UpdateModules: make(map[string]*rawYAMLConfig),
	}
}

type rawYAMLConfig struct {
	unmarshal func(interface{}) error
}

func (cfg *rawYAMLConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	cfg.unmarshal = unmarshal
	return nil
}

func (cfg *rawYAMLConfig) Unmarshal(v interface{}) error {
	return cfg.unmarshal(v)
}
