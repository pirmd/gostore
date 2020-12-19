package util

import (
	"testing"
)

func TestHasMajorChanges(t *testing.T) {
	tstCases := []struct {
		inL, inR map[string]interface{}
		want     bool
	}{
		{
			inL:  map[string]interface{}{"Title": "Hello World!"},
			inR:  map[string]interface{}{"Title": "Hello World!"},
			want: false,
		},
		{
			inL:  map[string]interface{}{"Title": "Hello World!"},
			inR:  map[string]interface{}{"Title": "Hello world"},
			want: false,
		},
		{
			inL:  map[string]interface{}{"Title": "Hello World!"},
			inR:  map[string]interface{}{"Title": "Hello folks!"},
			want: true,
		},
	}

	for _, tc := range tstCases {
		if got := HasMajorChanges(tc.inL, tc.inR); got != tc.want {
			t.Errorf("HasMajorChanges between %v and %v. Got: %v. Want: %v", tc.inL, tc.inR, got, tc.want)
		}
	}
}
