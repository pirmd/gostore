package googlebooks

import (
	"testing"
)

func TestParseTitle(t *testing.T) {
	testCases := []struct {
		in      *googleVolumeInfo
		outT    string
		outSubT string
		outS    string
		outN    string
	}{
		{&googleVolumeInfo{Title: "Sun Company (La compagnie des glaces 25)", SubTitle: "FNA 1431"}, "Sun Company", "FNA 1431", "La compagnie des glaces", "25"},
		{&googleVolumeInfo{Title: "Sun Company - La compagnie des glaces 25", SubTitle: "FNA 1431"}, "Sun Company", "FNA 1431", "La compagnie des glaces", "25"},
		{&googleVolumeInfo{Title: "Sun Company (La compagnie des glaces #25)", SubTitle: "FNA 1431"}, "Sun Company", "FNA 1431", "La compagnie des glaces", "25"},
		{&googleVolumeInfo{Title: "Sun Company (La compagnie des glaces n°25)", SubTitle: "FNA 1431"}, "Sun Company", "FNA 1431", "La compagnie des glaces", "25"},
		{&googleVolumeInfo{Title: "Sun Company (La compagnie des glaces Series 25)", SubTitle: "FNA 1431"}, "Sun Company", "FNA 1431", "La compagnie des glaces", "25"},
		{&googleVolumeInfo{Title: "Sun Company", SubTitle: "La compagnie des glaces 25"}, "Sun Company", "", "La compagnie des glaces", "25"},
		{&googleVolumeInfo{Title: "Sun Company", SubTitle: "Book 25 of La compagnie des glaces"}, "Sun Company", "", "La compagnie des glaces", "25"},
		{&googleVolumeInfo{Title: "Unknown"}, "Unknown", "", "", ""},
	}

	g := &googleBooks{}

	for _, tc := range testCases {
		gotT, gotSubT, gotS, gotN := g.parseTitle(tc.in)
		if gotT != tc.outT || gotSubT != tc.outSubT || gotS != tc.outS || gotN != tc.outN {
			t.Errorf("Guessing %#v failed:\nWant: %s, %s, %s, %s\nGot : %s, %s, %s, %s\n\n", tc.in, tc.outT, tc.outSubT, tc.outS, tc.outN, gotT, gotSubT, gotS, gotN)
		}
	}
}
