// Package formatter provides a register of formatting functions.  This
// register selects the proper formatting function according to the "type" of
// the object to format to string.  Type is guessed (in that order): by the
// TypeOf function if the object implement the Classfier interface, the
// FormatterTypeField (default to Type) if the object is a map or a struct or
// the golang object's type.
//
// Such a register of functions is helpful to quickly define a pretty printer.
package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

const (
	//DefaultFormatter is the name of default formatter that is used if the
	//provided attribute's type is not present in the register
	DefaultFormatter = "_default"
)

var (
	//ErrNoFormatterFound error is raised if no formating function was found in
	//the formatters register
	ErrNoFormatterFound = fmt.Errorf("formatter: no formatter found for the given type")
)

//Func represents a formating function
type Func func(v interface{}) (string, error)

//Formatters associates formating functions to a type
type Formatters map[string]Func

//Get retrieves the formating function corresponding to the given type.
//Should no formatter exists in the register for this type, the formatter
//registered for the DefaultFormatter is returned.
//
//ErrUnknownType is returned if no default formatter has been defined
func (f Formatters) Get(typ string) (Func, error) {
	if typ != "" {
		if fmtFn, exists := f[typ]; exists {
			return fmtFn, nil
		}
	}

	if fmtFn, exists := f[DefaultFormatter]; exists {
		return fmtFn, nil
	}

	return nil, ErrNoFormatterFound
}

//FormatUsingType applies the register's formatting function for the
//given type
func (f Formatters) FormatUsingType(typ string, v interface{}) (string, error) {
	fmtFn, err := f.Get(typ)
	if err != nil {
		return "", err
	}

	return fmtFn(v)
}

//Format applies the register's formatting function
func (f Formatters) Format(v interface{}) (string, error) {
	return f.FormatUsingType(TypeOf(v), v)
}

//MustFormat acts as Format but does not fail on error
//It captures any error to the output string, in case no formatter
//is found it reports golang string representation of of v
//(i.e. fmt.Printf("%+v", v))
func (f Formatters) MustFormat(v interface{}) string {
	s, err := f.Format(v)
	if err != nil {
		if err == ErrNoFormatterFound {
			return fmt.Sprintf("%+v", v)
		}

		return fmt.Sprintf("!Err(%s)", err)
	}

	return s
}

//MustFormatUsingType acts as FormatUsingType but does not fail on error
//It captures any error to the output string, in case no formatter
//is found it reports golang string representation of of v
//(i.e. fmt.Printf("%+v", v))
func (f Formatters) MustFormatUsingType(typ string, v interface{}) string {
	s, err := f.FormatUsingType(typ, v)
	if err != nil {
		if err == ErrNoFormatterFound {
			return fmt.Sprintf("%+v", v)
		}

		return fmt.Sprintf("!Err(%s)", err)
	}

	return s
}

//Register registers a new formatter in the formatters register
//It replaces any pre-existing formatter for the same type.
func (f Formatters) Register(forType string, fmtFn Func) {
	f[forType] = fmtFn
}

//JSONFormatter formats an interface using JSON marshaler
func JSONFormatter(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

//TemplateFormatter returns a Func based on a text/Template
//
//If provided text cannot be parsed as a text/template, the function panics
func TemplateFormatter(tmpl *template.Template, text string) Func {
	if _, err := tmpl.Parse(text); err != nil {
		panic(err)
	}

	return func(v interface{}) (string, error) {
		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, v); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}

//TemplateNewFormatter is an helper to quickly set-up a Func from a text/template string
func TemplateNewFormatter(text string) Func {
	return TemplateFormatter(new(template.Template), text)
}
