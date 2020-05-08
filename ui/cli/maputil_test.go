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

func TestGetCommonKeys(t *testing.T) {
	maps := []map[string]interface{}{
		{"a": "", "b": "", "c": ""},
		{"a": "", "b": "", "c": "", "d": ""},
		{"a": "", "c": ""},
	}

	tstCases := []struct {
		in   []string
		want []string
	}{
		{
			in:   []string{"a", "b", "c"},
			want: []string{"a", "c"},
		},
		{
			in:   []string{"a", "b", "*"},
			want: []string{"a", "c"},
		},
		{
			in:   []string{"c", "a", "*"},
			want: []string{"c", "a"},
		},

		{
			in:   []string{"!a", "b", "*", "c"},
			want: []string{"c"},
		},
	}

	for _, tc := range tstCases {
		got := getCommonKeys(maps, tc.in...)
		if failure := verify.Equal(got, tc.want); failure != nil {
			t.Errorf("Failed to select common fields in map collection:\n%v", failure)
		}
	}
}
