package formatter

import (
	"reflect"
)

var (
	//FormatterTypeField is the name of the attributes'map key
	//that determines the type of formatter to use
	FormatterTypeField = "Type"
)

//Classifier is an interface describing any object keen to disclose its identity
//It is helpful to select te proper Fomratter from a Formatters
type Classifier interface {
	Type() string
}

//TypeOf finds the type of a given object. Type is guessed (in that order):
//by the TypeOf function if the object implements the Classifier interface, the
//FormatterTypeField (default to Type) if the object is a map or a struct or
//the golang object's type representation.
func TypeOf(v interface{}) string {
	if classifier, ok := v.(Classifier); ok {
		return classifier.Type()
	}

	val := reflect.ValueOf(v)
	typ := val.Type()
	switch typ.Kind() {
	case reflect.Map:
		if typ.Key().Kind() == reflect.String {
			key := reflect.ValueOf(FormatterTypeField)
			entry := val.MapIndex(key)
			if entry.IsValid() {
				if t, ok := entry.Interface().(string); ok {
					return t
				}
			}
		}

	case reflect.Struct:
		field := val.FieldByName(FormatterTypeField)
		if field.IsValid() && field.CanInterface() {
			if t, ok := field.Interface().(string); ok {
				return t
			}
		}

	case reflect.Ptr:
		ptrElem := val.Elem()
		if ptrElem.IsValid() && ptrElem.CanInterface() {
			return TypeOf(ptrElem.Interface())
		}
	}

	return typ.String()
}
