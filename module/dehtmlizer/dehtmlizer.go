package dehtmlizer

import (
	"fmt"

	"github.com/pirmd/style"

	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "dehtmlizer"
)

var (
	_ module.Module = (*dehtmlizer)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
	// Fields lists the record's fields that should be cleaned from any html
	// tags.
	// Non-existing fields are silently ignored.
	Fields []string

	// OutputStyle is the name of the target style to use when converting html
	// to text. Known styles are text and markdown.
	OutputStyle string
}

func newConfig() module.Factory {
	return &config{}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newDehtmlizer(cfg, env)
}

// dehtmlizer is a gostore's module that converts text in html format to
// something more reasonable for metadata like pure text or markdown.
//
// dehtmlizer is pretty basic so only a sub-part of html formats are correctly
// interpreted.
type dehtmlizer struct {
	*module.Environment

	fields      []string
	outputStyle style.Styler
}

func newDehtmlizer(cfg *config, env *module.Environment) (*dehtmlizer, error) {
	d := &dehtmlizer{
		Environment: env,
		fields:      cfg.Fields,
		outputStyle: style.NewPlaintext(),
	}

	switch cfg.OutputStyle {
	case "": //Do nothing, keep defaults
	case "text":
		d.outputStyle = style.NewPlaintext()
	case "markdown":
		d.outputStyle = style.NewMarkdown()
	default:
		return nil, fmt.Errorf("module '%s': bad configuration: '%s' style is not supported", moduleName, cfg.OutputStyle)
	}

	return d, nil
}

// Process transforms any text in HTML format into markdown,
func (d *dehtmlizer) Process(r *store.Record) error {
	for _, field := range d.fields {
		d.Logger.Printf("Module '%s': clean field '%s'", moduleName, field)
		if err := d.html2txt(r, field); err != nil {
			return err
		}
	}
	return nil
}

func (d *dehtmlizer) html2txt(r *store.Record, field string) error {
	value := r.Get(field)
	if value == nil {
		return nil
	}

	html, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s: cannot dehtmlize field '%s' that does not contain text", moduleName, field)
	}

	txt, err := html2txt(html, d.outputStyle)
	if err != nil {
		return err
	}

	r.Set(field, txt)
	return nil
}

func init() {
	module.Register(moduleName, newConfig)
}
