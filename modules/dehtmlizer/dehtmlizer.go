// Package dehtmlizer converts text in html format to something more reasonable
// for metadata like pure text or markdown.
//
// dehtmlizer is pretty basic so only a sub-part of html formats are correctly
// interpreted.
package dehtmlizer

import (
	"fmt"
	"log"

	"github.com/pirmd/cli/style"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "dehtmlizer"
)

var (
	//Makes sure that dehtmlizer implements modules.Module
	_ modules.Module = (*dehtmlizer)(nil)
)

//Config defines the different configurations that can be used to customized
//the behavior of a dehtmlizer module.
type Config struct {
	//Fields2Clean lists the record's fields that should be dehtlmized.
	//Non-existing fields are ignored.
	Fields2Clean []string

	//OutputStyle is the name of the target style to use when converting html
	//to text. Known styles are text and markdown.
	OutputStyle string
}

func newConfig() *Config {
	return &Config{}
}

type dehtmlizer struct {
	log *log.Logger

	//TODO(pirmd): Allow configuration of which field to clean (maybe depending
	//on Record's Type like renamer)
	fields2clean []string

	outputStyle style.Styler
}

func newDehtmlizer(cfg *Config, logger *log.Logger) (modules.Module, error) {
	d := &dehtmlizer{
		log:         logger,
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

//ProcessRecord transforms any text in HTML format into markdown,
func (d *dehtmlizer) ProcessRecord(r *store.Record) error {
	for _, field := range d.fields2clean {
		if err := d.html2txt(r, field); err != nil {
			return err
		}
	}
	return nil
}

func (d *dehtmlizer) html2txt(r *store.Record, field string) error {
	value := r.GetValue(field)
	if value == nil {
		return nil
	}

	//TODO(pirmd): we are lead to make a lot of assumptions of the type of
	//stored attribute, need to do something better than map[string]interface{}
	html, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s: cannot dehtmlize field '%s' that does not contain text", moduleName, field)
	}

	txt, err := html2txt(html, d.outputStyle)
	if err != nil {
		return err
	}

	r.SetValue(field, txt)
	return nil
}

//New creates a new dehtmlizer module whose configuration information
func New(rawcfg modules.ConfigUnmarshaler, log *log.Logger) (modules.Module, error) {
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newDehtmlizer(cfg, log)
}

func init() {
	modules.Register(moduleName, New)
}
