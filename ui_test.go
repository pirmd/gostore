package main

import (
	"github.com/pirmd/verify"
	"testing"
)

func TestGetKeys(t *testing.T) {
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
			want: []string{"a", "b", "c"},
		},
		{
			in:   []string{"c", "a", "*"},
			want: []string{"c", "a", "b", "d"},
		},

		{
			in:   []string{"!a", "*", "c"},
			want: []string{"b", "d", "c"},
		},
	}

	for _, tc := range tstCases {
		got := getKeys(maps, tc.in...)
		verify.Equal(t, got, tc.want, "failed to select fields in map collection")
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
		verify.Equal(t, got, tc.want, "failed to select common fields in map collection")
	}
}
