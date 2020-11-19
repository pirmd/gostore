package dupfinder

import (
	"testing"
)

func TestEscapeQuery(t *testing.T) {
	testCases := []struct {
		in   string
		want string
	}{
		{"toto", "toto"},
		{"toto tutu", "toto\\ tutu"},
	}

	for _, tc := range testCases {
		got := escapeQuery(tc.in)

		if tc.want != got {
			t.Errorf("Fail to escape query %s.\nWant: %v\nGot : %v", tc.in, tc.want, got)
		}
	}
}
