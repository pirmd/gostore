package store

import (
	"reflect"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	testCases := []struct {
		in   string
		want time.Time
	}{
		{"1976", time.Date(1976, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"1976-01", time.Date(1976, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"1976-01-17", time.Date(1976, 1, 17, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range testCases {
		got, err := parseTime(tc.in)
		if err != nil {
			t.Errorf("Fail to parse time for %s: %v", tc.in, err)
		}

		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("Fail to parse time for %s.\nWant: %v\nGot : %v", tc.in, tc.want, got)
		}
	}
}
