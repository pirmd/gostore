package util

import (
	"reflect"
)

// IsZero checks whether the provided interface is the zero value.
func IsZero(v interface{}) bool {
	val := reflect.ValueOf(v)
	return !val.IsValid() || reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}
