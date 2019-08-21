package organizer

import (
	"fmt"
	"log"
	"path/filepath"
	"text/template"

	"github.com/pirmd/cli/formatter"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "organizer"
)

var (
	//Makes sure that organizer implements modules.Module
	_ modules.Module = (*organizer)(nil)

	//ErrEmptyName error is raised when the generated name is empty, main
	//reason is probably that the namer scheme does not match with the provided
	//attribute map (for example, lack of meaningful record's information like
	//book "Title").
	ErrEmptyName = fmt.Errorf("generated name is empty")
)

//Config defines the different configurations that can be used to customized
//the behavior of an organizer module.
type Config struct {
	//NamingSchemes defines, for each record's type, the templates to rename a
	//record according to its attribute.  You can define a default naming
	//scheme for all record's type not defined in NamingScheme using the
	//special "_default" type.
	NamingSchemes map[string]string

	//Sanitizer defines the name of the path sanitizer to use.
	//Available sanitizers are "none" (or ""), "standard", "nospace"
	Sanitizer string

	//XXX: see beets ou polochon to get inspired by names (NamingScheme + Sanitizer)
}

type organizer struct {
	log *log.Logger

	namers    formatter.Formatters
	sanitizer func(string) string
}

//New creates a new organizer module
func New(config interface{}, logger *log.Logger) (modules.Module, error) {
	o := &organizer{
		namers:    formatter.Formatters{},
		sanitizer: func(s string) string { return s },
		log:       logger,
	}

	cfg, ok := config.(Config)
	if !ok {
		return nil, fmt.Errorf("%s: wrong configuration format", moduleName)
	}

	tmpl := template.New("organizer").Funcs(map[string]interface{}{
		//XXX: Provides helpers for naming scheme definition (?)
	})
	for typ, txt := range cfg.NamingSchemes {
		fmtFn := formatter.TemplateFormatter(tmpl.New(typ), txt)
		o.namers.Register(typ, fmtFn)
	}

	switch cfg.Sanitizer {
	case "", "none": // do nothing
	case "standard":
		o.sanitizer = pathSanitizer
	case "nospace":
		o.sanitizer = nospaceSanitizer
	default:
		return nil, fmt.Errorf("%s: sanitizer '%s' is unknown (can be none, standard or nospace)", moduleName, cfg.Sanitizer)
	}

	return o, nil
}

//ProcessRecord modifies the record's name to match a standardized naming scheme.
func (o *organizer) ProcessRecord(r *store.Record) error {
	name, err := o.namers.Format(r.Fields())
	if err != nil {
		return err
	}

	if filepath.Base(name) == "" {
		return ErrEmptyName
	}

	if filepath.Ext(name) == "" {
		name = name + filepath.Ext(r.Key())
	}

	//name should be relative to the collection's root, clean
	//unuseful cruft.
	name = filepath.ToSlash(filepath.Clean("/" + name))[1:]

	r.SetKey(name)
	return nil
}

func init() {
	modules.Register(moduleName, New)
}
