package books

import (
	"testing"
)

func TestGuessSerie(t *testing.T) {
	testCases := []struct {
		in   string
		outT string
		outS string
		outN string
	}{
		{"Sun Company (La compagnie des glaces 25)", "Sun Company", "La compagnie des glaces", "25"},
		{"Sun Company - La compagnie des glaces 25", "Sun Company", "La compagnie des glaces", "25"},
		{"Sun Company (La compagnie des glaces #25)", "Sun Company", "La compagnie des glaces", "25"},
		{"Sun Company (La compagnie des glaces n°25)", "Sun Company", "La compagnie des glaces", "25"},
		{"Sun Company (La compagnie des glaces Series 25)", "Sun Company", "La compagnie des glaces", "25"},
		{"Book 25 of La compagnie des glaces", "", "La compagnie des glaces", "25"},
		{"Unknown", "Unknown", "", ""},
	}

	for _, tc := range testCases {
		gotT, gotS, gotN := GuessSerie(tc.in)
		if gotT != tc.outT || gotS != tc.outS || gotN != tc.outN {
			t.Errorf("Guessing %#v failed:\nWant: %s, %s, %s\nGot : %s, %s, %s\n\n", tc.in, tc.outT, tc.outS, tc.outN, gotT, gotS, gotN)
		}
	}
}
