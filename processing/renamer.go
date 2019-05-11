package processing

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/pirmd/cli/formatter"
	"github.com/pirmd/gostore/store"
)

var (
	//ErrEmptyName error is raised when the generated name is empty, main
	//reason is probably that the namer scheme does not match with the provided
	//attribute map (for example, lack of meaningful key/values like book "Title")
	ErrEmptyName = fmt.Errorf("namer: generated name is empty")

	namerTmpl = template.New("namer")
	namers    = formatter.Formatters{}
)

func AddNamer(name string, text string) {
	namers.Register(name, formatter.TemplateFormatter(namerTmpl.New(name), text))
}

//RenameRecord proposes a standradized name build after the provided
//attributes map. To generate the name, a namer function corresponding
//to the map type is looked for in the namers register
func RenameRecord(r *store.Record) error {
	name, err := namers.Format(r.Fields())
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
	RecordProcessors["rename"] = RenameRecord
}
