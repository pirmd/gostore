package store

import (
	"fmt"
	"io"
	"time"

	"github.com/pirmd/gostore/util"
)

const (
	// KeyField contains the name of the record's value field containing the
	// record's key when exported through Fields()
	KeyField = "Name"
)

// Record represents a Store's record.
type Record struct {
	*value
	key  string
	file ReadCloser
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

// Flatted returns all Record's data in a single flat map including Record's Key
func (r *Record) Flatted() map[string]interface{} {
	flatted := r.Value()
	flatted[KeyField] = r.key
	return flatted
}

// File returns the file attached to a Record. It rewinds the Record's file to
// its start before returning it.  File is nil if Record's as no known attached
// file.  File is usually needed to reference source media file of a record
// before having it inserted or updated in the store.
func (r *Record) File() ReadCloser {
	// take benefit of r.file being a Seeker to rewind to the start and allow
	// several chain reading of r (to chain modules for example)
	if r.file != nil {
		r.file.Seek(0, io.SeekStart)
	}

	return r.file
}

// SetFile sets Record's file
func (r *Record) SetFile(f ReadCloser) {
	r.file = f
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

// ReadCloser is an interface that wraps all io Read methods that are usually
// need to play with media files.
type ReadCloser interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

// value represents a set of information recorded in the store
type value struct {
	// CreatedAt is the time stamp corresponding to the record's creation date
	CreatedAt time.Time
	// UpdatedAt is the time stamp corresponding to the last record's update
	UpdatedAt time.Time
	// UserData is a dictionary of all user-supplied data stored in the record
	UserData map[string]interface{}
}

// newValue creates a new value
func newValue(data map[string]interface{}) *value {
	val := &value{
		CreatedAt: timestamper(),
		UpdatedAt: timestamper(),
		UserData:  make(map[string]interface{}),
	}
	val.SetData(data)
	return val
}

// Data returns stored user-supplied data
func (val *value) Data() map[string]interface{} {
	return util.CopyMap(val.UserData)
}

// SetData replaces record's user-supplied data.
func (val *value) SetData(data map[string]interface{}) {
	val.UserData = make(map[string]interface{})
	val.MergeData(data)
}

// MergeData completes and replaces stored data with data content (data values
// are copied and supersede record's existing values. Record's values not in
// data are kept as is).
func (val *value) MergeData(data map[string]interface{}) {
	for k, v := range data {
		val.UserData[k] = v
	}

	val.UpdatedAt = timestamper()
}

// Value returns all information stored in value, both user-supplied data
// and automatic managed data like creation/update time.
func (val *value) Value() map[string]interface{} {
	v := val.Data()
	v["CreatedAt"] = val.CreatedAt
	v["UpdatedAt"] = val.UpdatedAt
	return v
}

// Get retrieves a (key, value) information stored in a Value.
func (val *value) Get(key string) interface{} {
	return val.UserData[key]
}

// Set adds a new (key, value).
func (val *value) Set(k string, v interface{}) {
	if fmt.Sprint(val.UserData[k]) == fmt.Sprint(v) {
		return
	}

	val.UserData[k] = v
	val.UpdatedAt = timestamper()
}

// SetIfExists updates a value if already exists.
func (val *value) SetIfExists(k string, v interface{}) {
	if _, exists := val.UserData[k]; exists {
		val.Set(k, v)
	}
}

// Del removes a (key, value).
func (val *value) Del(k string) {
	if _, exists := val.UserData[k]; exists {
		delete(val.UserData, k)
		val.UpdatedAt = timestamper()
	}
}
