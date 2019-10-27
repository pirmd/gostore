package media_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/pirmd/gostore/media"
	_ "github.com/pirmd/gostore/media/books"
	"github.com/pirmd/verify"
)

const (
	testdataPath = "../testdata" //Use test data of the main gostore package
)

func TestGetMetadata(t *testing.T) {
	testCases, err := filepath.Glob(filepath.Join(testdataPath, "*.epub"))
	if err != nil {
		t.Fatalf("cannot read test data in %s:%v", testdataPath, err)
	}

	out := []media.Metadata{}
	for _, tc := range testCases {
		m, err := media.GetMetadataFromFile(tc)
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
