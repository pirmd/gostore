package organizer

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"text/template"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "organizer"
)

var (
	// Makes sure that organizer implements modules.Module
	_ modules.Module = (*organizer)(nil)

	// DefaultNamingScheme is the name of the default NamingScheme
	DefaultNamingScheme = media.DefaultType

	// ErrNoNamingScheme is raised when no naming scheme is found, even
	// DefaultNamingScheme
	ErrNoNamingScheme = errors.New(moduleName + ": no naming scheme found")

	// ErrEmptyName error is raised when the generated name is empty, main
	// reason is probably that the namer scheme does not match with the provided
	// attribute map (for example, lack of meaningful record's information like
	// book "Title").
	ErrEmptyName = errors.New(moduleName + ": generated name is empty")
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
			DefaultNamingScheme: "{{ .Name }}",
		},
	}
}

type organizer struct {
	log *log.Logger

	namers *template.Template
}

func newOrganizer(cfg *Config, logger *log.Logger) (*organizer, error) {
	o := &organizer{
		log:    logger,
		namers: template.New("organizer"),
	}
	o.namers.Funcs(o.funcmap())

	for typ, txt := range cfg.NamingSchemes {
		if _, err := o.namers.New(typ).Parse(txt); err != nil {
			return nil, err
		}
	}

	return o, nil
}

// ProcessRecord modifies the record's name to match a standardized naming scheme.
func (o *organizer) ProcessRecord(r *store.Record) error {
	name, err := o.name(r.Flatted())
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

func (o *organizer) name(m map[string]interface{}) (string, error) {
	t := o.namerFor(m)
	if t == nil {
		return "", ErrNoNamingScheme
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, m); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (o *organizer) namerFor(m map[string]interface{}) *template.Template {
	typ := media.TypeOf(m)

	if tmpl := o.namers.Lookup(typ); tmpl != nil {
		return tmpl
	}

	if tmpl := o.namers.Lookup(filepath.Base(typ)); tmpl != nil {
		return tmpl
	}

	if tmpl := o.namers.Lookup(filepath.Dir(typ)); tmpl != nil {
		return tmpl
	}

	return o.namers.Lookup(DefaultNamingScheme)
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
