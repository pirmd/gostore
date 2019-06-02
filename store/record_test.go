package store

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestValueSet(t *testing.T) {
	testCases := []struct {
		inKey, inVal string
		want         Value
	}{
		{"created", "2018-11-11", Value{"created": "2018-11-11"}},
		{"tstDate", "2018-11-11", Value{"tstDate": time.Date(2018, 11, 11, 0, 0, 0, 0, time.UTC)}},
		{CreatedAtField, "2018-11-11", Value{CreatedAtField: time.Date(2018, 11, 11, 0, 0, 0, 0, time.UTC)}},
	}

	for _, tc := range testCases {
		got := make(Value)
		got.Set(tc.inKey, tc.inVal)
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("Fail to set value with correct type.\nWant: %#v\nGot : %#v", tc.want, got)
		}
	}
}

func TestValueFromJson(t *testing.T) {
	testCases := []struct {
		in   string
		want Value
	}{
		{`{"created":  "2018-11-11"}`, Value{"created": "2018-11-11"}},
		{`{"tstDate":  "2018-11-11"}`, Value{"tstDate": time.Date(2018, 11, 11, 0, 0, 0, 0, time.UTC)}},
		{`{"CreatedAt":"2018-11-11"}`, Value{"CreatedAt": time.Date(2018, 11, 11, 0, 0, 0, 0, time.UTC)}},
		{`{"SerieIndex":"2.0"}`, Value{"SerieIndex": 2.0}},
	}

	for _, tc := range testCases {
		got := make(Value)
		if err := json.Unmarshal([]byte(tc.in), &got); err != nil {
			t.Errorf("Failed to unmarshal %s: %s", tc.in, err)
		}

		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("Fail to unmarshal %s.\nWant: %#v\nGot : %#v", tc.in, tc.want, got)
		}
	}
}
