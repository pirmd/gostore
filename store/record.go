package store

import (
	"fmt"
	"time"
)

const (
	// KeyField contains the name of the record's value field containing the
	// record's key when exported through Fields()
	KeyField = "Name"
)

// Record represents a Store's record.
type Record struct {
	key   string
	value *value
}

// NewRecord creates a Record.
func NewRecord(key string, data map[string]interface{}) *Record {
	return &Record{
		key:   key,
		value: newValue(data),
	}
}

// Key is Record's (unique) identifier in the store
func (r *Record) Key() string {
	return r.key
}

// SetKey modifies Record's (unique) identifier
func (r *Record) SetKey(key string) {
	r.key = key
}

// Value returns a copy of all information known about Record. It contains the
// information supplied by the end-user as well as information auto-generated
// during Record's management (like creation/update stamps).
func (r *Record) Value() map[string]interface{} {
	return r.value.Flatted()
}

// Data returns a copy of the end-user supplied information, information
// auto-managed by the Record are filtered out.
func (r *Record) Data() map[string]interface{} {
	return r.value.GetData()
}

// SetData replaces Record's content with the given information.
func (r *Record) SetData(data map[string]interface{}) {
	r.value.SetData(data)
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
func (r *Record) Set(k string, v interface{}) {
	r.value.Set(k, v)
}

// SetIfExists updates a record's stored information if already exists.
func (r *Record) SetIfExists(k string, v interface{}) {
	r.value.SetIfExists(k, v)
}

// Del removes a record's stored information.
func (r *Record) Del(k string) {
	r.value.Del(k)
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

// Data returns the Data() for each record in the collection.
func (r Records) Data() (v []map[string]interface{}) {
	for _, i := range r {
		v = append(v, i.Data())
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

// value represents a set of information recorded in the store
type value struct {
	// CreatedAt is the time stamp corresponding to the record's creation date
	CreatedAt time.Time
	// UpdatedAt is the time stamp corresponding to the last record's update
	UpdatedAt time.Time
	// Data is a dictionary of all user-supplied data stored in the record
	Data map[string]interface{}
}

// newValue creates a new value
func newValue(data map[string]interface{}) *value {
	val := &value{
		CreatedAt: timestamper(),
	}
	val.SetData(data)
	return val
}

// SetData replaces user-supplied value.
func (val *value) SetData(data map[string]interface{}) {
	if fmt.Sprint(val.Data) == fmt.Sprint(data) {
		return
	}

	val.Data = make(map[string]interface{})
	for k, v := range data {
		val.Set(k, v)
	}
	val.UpdatedAt = timestamper()
}

// GetData returns stored data
func (val *value) GetData() map[string]interface{} {
	data := make(map[string]interface{})
	for k, v := range val.Data {
		data[k] = v
	}
	return data
}

// Flatted returns all information stored in value, both user-supplied data
// and automatic managed data like creation/update time.
func (val *value) Flatted() map[string]interface{} {
	flatted := val.GetData()
	flatted["CreatedAt"] = val.CreatedAt
	flatted["UpdatedAt"] = val.UpdatedAt
	return flatted
}

// Get retrieves a (key, value) information stored in a Value.
func (val *value) Get(key string) interface{} {
	return val.Data[key]
}

// Set adds a new (key, value).
func (val *value) Set(k string, v interface{}) {
	val.Data[k] = v
}

// SetIfExists updates a value if already exists.
func (val *value) SetIfExists(k string, v interface{}) {
	if _, exists := val.Data[k]; exists {
		val.Data[k] = v
	}
}

// Del removes a (key, value).
func (val *value) Del(k string) {
	delete(val.Data, k)
}
