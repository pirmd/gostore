package store

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	// KeyField contains the name of the record's value field containing the
	// record's key when exported through Fields()
	KeyField = "Name"

	// CreatedAtField contains the name of the record's value field containing
	// the time stamp corresponding to the record's creation
	CreatedAtField = "CreatedAt"

	// UpdatedAtField contains the name of the record's value field containing
	// the time stamp corresponding to the last known record's update
	UpdatedAtField = "UpdatedAt"
)

var (
	// names of fields that are auto-managed by the Record struct
	autoManagedFields = []string{CreatedAtField, UpdatedAtField}
)

// Record represents a Store's record.
type Record struct {
	key   string
	value Value
}

// NewRecord creates a Record.
func NewRecord(key string, value map[string]interface{}) *Record {
	r := &Record{key, make(Value)}
	for k, v := range value {
		r.value.Set(k, v)
	}
	r.stamp()
	return r
}

func (r *Record) String() string {
	return fmt.Sprintf("%s:%#v", r.key, r.value)
}

// Key is Record's (unique) identifier in the store
func (r *Record) Key() string {
	return r.key
}

// SetKey modifies Record's (unique) identifier
func (r *Record) SetKey(key string) {
	r.key = key
	r.stamp()
}

// Value returns a copy of all information known about Record. It contains the
// information supplied by the end-user as well as information auto-generated
// during Record's management (like creation/update stamps).
func (r *Record) Value() map[string]interface{} {
	val := make(map[string]interface{})
	for k, v := range r.value {
		val[k] = v
	}
	return val
}

// SetValue set/updates record's information.
func (r *Record) SetValue(m map[string]interface{}) {
	for k, v := range m {
		r.value.Set(k, v)
	}
	r.stamp()
}

// ReplaceValue replaces Record's content with the given information.
// Initial CreatedAtField value is kept if not explicitly asked to be replaced.
func (r *Record) ReplaceValue(fields map[string]interface{}) {
	createdAt := r.value.Get(CreatedAtField)

	r.value = make(Value)
	for k, v := range fields {
		r.value.Set(k, v)
	}

	r.value.SetIfNotExists(CreatedAtField, createdAt)
	r.stamp()
}

// UserValue returns a copy of the end-user supplied information, information auto-managed by the Record are filtered out.
func (r *Record) UserValue() map[string]interface{} {
	orig := r.Value()
	for _, k := range autoManagedFields {
		delete(orig, k)
	}
	return orig
}

// Flatted returns all Record's data in a single flat map including Record's Key
func (r *Record) Flatted() map[string]interface{} {
	flatted := r.Value()
	flatted[KeyField] = r.key
	return flatted
}

// Get retrieves a Record's stored information.
func (r *Record) Get(k string) interface{} {
	return r.value.Get(k)
}

// Set adds/modifies a record's stored information.
// Set ensures that fields supposed to host a time stamp or a date are of time type.
func (r *Record) Set(k string, v interface{}) {
	r.value.Set(k, v)
	r.stamp()
}

// stamp creates or updates "CreatedAtField" and "UpdatedAtField"
func (r *Record) stamp() {
	r.value.SetIfNotExists(CreatedAtField, timestamper())
	r.value.Set(UpdatedAtField, timestamper())
}

// Records represents a collection of Record
type Records []*Record

// Key returns the Key() for each record in the collection.
func (r Records) Key() (k []string) {
	for _, i := range r {
		k = append(k, i.Key())
	}
	return
}

// Value returns the Value() for each record in the collection.
func (r Records) Value() (v []map[string]interface{}) {
	for _, i := range r {
		v = append(v, i.Value())
	}
	return
}

// UserValue returns the Value() for each record in the collection.
func (r Records) UserValue() (v []map[string]interface{}) {
	for _, i := range r {
		v = append(v, i.UserValue())
	}
	return
}

// Flatted returns the Flatted() form for each record in the collection.
func (r Records) Flatted() (v []map[string]interface{}) {
	for _, i := range r {
		v = append(v, i.Flatted())
	}
	return
}

// Value represents a set of information recorded in the store
type Value map[string]interface{}

// Get retrieves a (key, value) information stored in a Value.
func (val Value) Get(key string) interface{} {
	return val[key]
}

// Set adds a new (key, value). It tries to ensure that:
// - time-stamps or dates are of time type. Recognized time-stamps or date are
// for CreatedAtField, UpdatedAtField or any field ending with "edAt" or "Date"
// keyword.
// - indexes, rates are of float type. Recognized indexes or rate are any field
// ending by "Index", "Position", "Rate"
// If it fails to parse a time or float, it falls back to a string.
func (val Value) Set(key string, value interface{}) {
	if _, ok := value.(string); !ok {
		val[key] = value
		return
	}

	if key == CreatedAtField || key == UpdatedAtField ||
		strings.HasSuffix(key, "Date") || strings.HasSuffix(key, "edAt") {
		if t, err := parseTime(value.(string)); err == nil {
			val[key] = t
			return
		}
		//TODO: do something if date is not readable
	}

	if strings.HasSuffix(key, "Index") || strings.HasSuffix(key, "Position") ||
		strings.HasSuffix(key, "Rate") {
		if f, err := strconv.ParseFloat(value.(string), 32); err == nil {
			val[key] = f
			return
		}
	}

	val[key] = value
}

// SetIfNotZero adds a new (key, value) only if the value is not Zero.
func (val Value) SetIfNotZero(key string, value interface{}) {
	if !isZero(value) {
		val.Set(key, value)
	}
}

// SetIfNotExists adds a new (key, value) only if no value already exists for
// the provided key
func (val Value) SetIfNotExists(key string, value interface{}) {
	if _, exists := val[key]; !exists {
		val.Set(key, value)
	}
}

// UnmarshalJSON personalizes records retrieving from store It mainly detects
// fields supposed to handle time-stamps or date. If it fails to parse a time,
// it falls back to a string
func (val Value) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	for k, v := range raw {
		val.Set(k, v)
	}
	return nil
}

func isZero(v interface{}) bool {
	val := reflect.ValueOf(v)
	return !val.IsValid() || reflect.DeepEqual(val.Interface(), reflect.Zero(val.Type()).Interface())
}
