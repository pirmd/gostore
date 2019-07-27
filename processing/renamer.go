package processing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

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

//AddRenamer register a new namer
func AddRenamer(name string, text string) {
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

//DetoxRecordName replaces some unreasable rune from a record's name in the
//hope to sore it with a cleaned filename. It borrows some basic rules from
//detox tool.
func DetoxRecordName(r *store.Record) error {
	r.SetKey(pathSanitizer(r.Key()))
	return nil
}

//pathSanitizer filters some unreasonable rune from a path name
func pathSanitizer(path string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) ||
			r == '_' || r == '-' ||
			r == '.' || r == os.PathSeparator {
			return r
		}
		if unicode.IsSpace(r) {
			return '_'
		}
		if unicode.In(r, unicode.Hyphen) || r == '\'' {
			return '-'
		}
		return -1
	}, path)
}

func init() {
	RecordProcessors["rename"] = RenameRecord
	RecordProcessors["detox"] = DetoxRecordName
}
