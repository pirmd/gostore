package formatter

import (
	"testing"
)

type test struct{}

func (t *test) Type() string { return "test" }

func TestTypeOf(t *testing.T) {
	testCases := []struct {
		in  interface{}
		out string
	}{
		{t, "testing.T"},
		{new(Formatters), "formatter.Formatters"},
		{&test{}, "test"},
		{map[string]string{"Type": "test"}, "test"},
		{map[string]interface{}{"Type": "test"}, "test"},
		{&struct{ Type string }{Type: "test"}, "test"},
	}

	for _, tc := range testCases {
		got := TypeOf(tc.in)
		if got != tc.out {
			t.Errorf("Fail to find correct Type for %+v.\nGot   : %s\nExpected: %s", tc.in, got, tc.out)
		}
	}
}
