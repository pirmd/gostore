package cli

import (
	"testing"

	"github.com/pirmd/verify"
)

func TestGetKeys(t *testing.T) {
	maps := []map[string]interface{}{
		{"a": "", "b": "", "c": ""},
		{"a": "A", "b": "", "c": "", "d": "D"},
		{"a": "", "c": ""},
	}

	tstCases := []struct {
		in   []string
		want []string
	}{
		{
			in:   []string{"a", "b", "c"},
			want: []string{"a", "b", "c"},
		},
		{
			in:   []string{"c", "a", "*"},
			want: []string{"c", "a", "b", "d"},
		},
		{
			in:   []string{"?a", "b", "c"},
			want: []string{"?a", "b", "c"},
		},
		{
			in:   []string{"c", "a", "?*"},
			want: []string{"c", "a", "?d"},
		},
		{
			in:   []string{"!a", "*", "c"},
			want: []string{"b", "d", "c"},
		},
	}

	for _, tc := range tstCases {
		got := getKeys(maps, tc.in...)
		if failure := verify.Equal(got, tc.want); failure != nil {
			t.Errorf("Failed to select fields in map collection:\n%v", failure)
		}
	}
}

func TestHasChanged(t *testing.T) {
	tstCases := []struct {
		inL, inR map[string]interface{}
		want     changeLevel
	}{
		{
			inL:  map[string]interface{}{"Title": "Hello World!"},
			inR:  map[string]interface{}{"Title": "Hello World!"},
			want: noChange,
		},
		{
			inL:  map[string]interface{}{"Title": "Hello World!"},
			inR:  map[string]interface{}{"Title": "Hello world"},
			want: minorChange,
		},
		{
			inL:  map[string]interface{}{"Title": "Hello World!"},
			inR:  map[string]interface{}{"Title": "Hello folks!"},
			want: majorChange,
		},
	}

	for _, tc := range tstCases {
		if got := hasChanged(tc.inL, tc.inR); got != tc.want {
			t.Errorf("Change between %v and %v. Got: %s. Want: %s", tc.inL, tc.inR, got, tc.want)
		}
	}
}
