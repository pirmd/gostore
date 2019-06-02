package books

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/verify"
)

const (
	testdataPath = "../../testdata" //Use test data of the main gostore package
)

func TestGetMetadata(t *testing.T) {
	testCases, err := filepath.Glob(filepath.Join(testdataPath, "*.epub"))
	if err != nil {
		t.Fatalf("cannot read test data in %s:%v", testdataPath, err)
	}

	epubH := &epubHandler{}

	out := []media.Metadata{}
	for _, tc := range testCases {
		f, err := os.Open(tc)
		if err != nil {
			t.Errorf("Failed to open test file %s: %v", tc, err)
		}
		defer f.Close()

		m, err := epubH.GetMetadata(f)
		if err != nil {
			t.Errorf("Fail to get metadata for %s: %v", tc, err)
		}
		out = append(out, m)
	}

	got, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatalf("Fail to marshal test output to json: %v", err)
	}

	verify.MatchGolden(t, string(got), "Metadata is not as expected.")
}
