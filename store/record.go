package store

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
    //Name of the record's value field containing the record's key when
    //exported through Fields()
	KeyField = "Name"
    //Name of the record's value field containint the time stamp corresponding
    //to the record's creation
	CreatedAtField = "CreatedAt"
    //Name of the record's value field containint the time stamp corresponding
    //to the last known record's update
	UpdatedAtField = "UpdatedAt"
)

type mdata map[string]interface{}

//Add adds a new (key, value) to mdata, it tries to ensure that timestamps are
//of time type.  Recognized time stamps should be in RFC3339 format, they are
//for CreatedAtField, UpdatedAtField or any field ending with "Date" keyword.
//If it fails to parse a time, it falls back to a string
func (m mdata) Add(key string, value interface{}) {
	if _, ok := value.(string); !ok {
		m[key] = value
		return
	}

	if key == CreatedAtField || key == UpdatedAtField || strings.HasSuffix(key, "Date") {
		//TODO: use a wider range of time formats (see http.Parsetime)
		if t, err := time.Parse(time.RFC3339, value.(string)); err == nil {
			m[key] = t
			return
		}
	}

	m[key] = value
}

//UnmarshalJSON personnalizes records retrieving from store It mainly detects
//fields supposed to handle time-stamps or date If it fails to parse a time, it
//falls back to a string
func (m mdata) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	for k, v := range raw {
		m.Add(k, v)
	}
	return nil
}

//Record represents a Store's record.
type Record struct {
	key   string
	value mdata
}

//NewRecord creates a Record.
func NewRecord(key string, value map[string]interface{}) *Record {
	r := &Record{key, value}
	if r.value == nil {
		r.value = make(mdata)
	}

	r.stamp()

	return r
}

func (r *Record) String() string {
	return fmt.Sprintf("%s:%#v", r.key, r.value)
}

//Key is Record's identifier in the store
func (r *Record) Key() string {
	return r.key
}

//SetKey modifies Record's identifier
func (r *Record) SetKey(key string) {
	r.key = key
	r.stamp()
}

//Value returns a copy of all information stored about Record.  It contains the
//information supplied by the user at Record's creation and auto-generated
//information like creation/update stamps.
func (r *Record) Value() map[string]interface{} {
	val := make(map[string]interface{})
	for k, v := range r.value {
		val[k] = v
	}
	return val
}

//OrigValue returns Record's values that are not managed by the store
func (r *Record) OrigValue() map[string]interface{} {
	orig := r.Value()
	for _, k := range []string{CreatedAtField, UpdatedAtField} {
		delete(orig, k)
	}
	return orig
}

//Fields returns all Record's attributes in a single flat map including the key
//value
func (r *Record) Fields() map[string]interface{} {
	fields := r.Value()
	fields[KeyField] = r.key
	return fields
}

//SetValue add/modified a record's value SetValue ensures that fields supposed
//to host a time stamp or a date are of time type.  SetValue creates or updates
//"CreatedAtField" and "UpdatedAtField"
func (r *Record) SetValue(k string, v interface{}) {
	r.value.Add(k, v)
	r.stamp()
}

//MergeValues updates record with the given fields' values.  MergeValues
//ensures that fields supposed to host a time stamp or a date are of time type.
//MergeValues creates or updates "CreatedAtField" and "UpdatedAtField"
func (r *Record) MergeValues(fields map[string]interface{}) {
	for k, v := range fields {
		r.value.Add(k, v)
	}
	r.stamp()
}

//ReplaceValues replaces record's values with the given fields CreatedAtField
//value is kept if not explicitly asked to be replaced ReplaceValues ensures
//that fields supposed to host a time stamp or a date are of time type.
//ReplaceValues creates or updates "CreatedAtField" and "UpdatedAtField"
func (r *Record) ReplaceValues(fields map[string]interface{}) {
	createdAt := r.value[CreatedAtField]

	r.value = make(mdata)
	for k, v := range fields {
		r.value.Add(k, v)
	}

	if _, exists := r.value[CreatedAtField]; !exists {
		r.value.Add(CreatedAtField, createdAt)
	}

	r.stamp()
}

//stamp creates or updates "CreatedAtField" and "UpdatedAtField"
func (r *Record) stamp() {
	if _, exists := r.value[CreatedAtField]; !exists {
		r.value[CreatedAtField] = time.Now()
	}
	r.value[UpdatedAtField] = time.Now()
}

//Records represents a collection of Record
type Records []*Record

//Key returns for each record in the collection their key
func (r Records) Key() (k []string) {
	for _, i := range r {
		k = append(k, i.Key())
	}
	return
}

//Value returns for each record in the collection their value
func (r Records) Value() (v []map[string]interface{}) {
	for _, i := range r {
		v = append(v, i.Value())
	}
	return
}

//Fields returns for each record in the collection their value and key in a
//single flat map
func (r Records) Fields() (v []map[string]interface{}) {
	for _, i := range r {
		v = append(v, i.Fields())
	}
	return
}
