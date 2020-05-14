package organizer

import (
	"fmt"
	"log"
	"path/filepath"
	"text/template"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui/formatter"
)

const (
	moduleName = "organizer"
)

var (
	// Makes sure that organizer implements modules.Module
	_ modules.Module = (*organizer)(nil)

	// ErrEmptyName error is raised when the generated name is empty, main
	// reason is probably that the namer scheme does not match with the provided
	// attribute map (for example, lack of meaningful record's information like
	// book "Title").
	ErrEmptyName = fmt.Errorf("generated name is empty")
)

// Config defines the different configurations that can be used to customize
// the behavior of an organizer module.
type Config struct {
	// NamingSchemes defines, for each record's type, the templates to rename a
	// record according to its attribute.  You can define a default naming
	// scheme for all record's type not defined in NamingScheme using the
	// special "_default" type.
	NamingSchemes map[string]string
}

func newConfig() *Config {
	return &Config{
		NamingSchemes: map[string]string{
			"_default": "{{ .Name }}",
		},
	}
}

type organizer struct {
	log *log.Logger

	namers formatter.Formatters
}

func newOrganizer(cfg *Config, logger *log.Logger) (*organizer, error) {
	o := &organizer{
		namers: formatter.Formatters{},
		log:    logger,
	}

	tmpl := template.New("organizer")
	tmpl.Funcs(funcmap(tmpl))
	for typ, txt := range cfg.NamingSchemes {
		fmtFn := formatter.TemplateFormatter(tmpl.New(typ), txt)
		o.namers.Register(typ, fmtFn)
	}

	return o, nil
}

// ProcessRecord modifies the record's name to match a standardized naming scheme.
func (o *organizer) ProcessRecord(r *store.Record) error {
	name, err := o.namers.Format(r.Fields())
	if err != nil {
		return err
	}

	if filepath.Base(name) == "" {
		return ErrEmptyName
	}

	//name should be relative to the collection's root, clean
	//unuseful cruft.
	name = filepath.ToSlash(filepath.Clean("/" + name))[1:]

	r.SetKey(name)
	return nil
}

// New creates a new organizer module
func New(rawcfg modules.ConfigUnmarshaler, log *log.Logger) (modules.Module, error) {
	log.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newOrganizer(cfg, log)
}

func init() {
	modules.Register(moduleName, New)
}
