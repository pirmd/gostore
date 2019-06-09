package store

import (
	"path/filepath"
	"testing"

	"github.com/blevesearch/bleve/analysis/lang/fr"
	"github.com/pirmd/verify"
)

func setupIdx(tb testing.TB) (*storeidx, func()) {
	tstDir := verify.NewTestField(tb)

	idx := newIdx(filepath.Join(tstDir.Root, "idxtest"))
	if err := idx.Open(); err != nil {
		tstDir.Clean()
		tb.Fatalf("Fail to create testing Store index: %s", err)
	}

	return idx, func() {
		idx.Close()
		tstDir.Clean()
	}
}

func populateIdx(tb testing.TB, idx *storeidx) (keys []string) {
	for _, td := range testData {
		r := NewRecord(buildKey(td), td)
		if err := idx.Put(r); err != nil {
			tb.Fatalf("Fail to add %v", td)
		}
		keys = append(keys, r.key)
	}
	return
}

func TestIndexSearch(t *testing.T) {
	idx, cleanFn := setupIdx(t)
	defer cleanFn()

	idx.Mapping.DefaultAnalyzer = fr.AnalyzerName
	keys := populateIdx(t, idx)

	testCases := []struct {
		name string
		in   string
		out  []string
	}{
		{"simple", "Charles-Michel de l'Épée", []string{keys[12], keys[5]}},
		{"simple", "Victor", []string{keys[2], keys[3]}},
		{"simple", "Michel de l'Épée", []string{keys[12], keys[5]}},
		{"case", "Nettoyage", []string{keys[3]}},
		{"case", "PÈRE", []string{keys[1]}},
		{"accent", "misérables", []string{keys[2]}},
		{"accent", "miserables", []string{keys[2]}},
		{"accent", "élément", []string{keys[0]}},
		{"accent", "elément", []string{keys[0]}},
		{"punctuation", "ce héros, mon père", []string{keys[1]}},
		{"punctuation", "Charles Michel", []string{keys[12]}},
		{"right order", "de l'Épée Charles-Michel", []string{keys[12], keys[5]}},
		{"right order", "l'Épée Charles-Michel (de)", []string{keys[12], keys[5]}},
		{"stemming", "misérable", []string{keys[2]}},
		{"stemming", "miserable", []string{keys[2]}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := idx.Search(tc.in)
			if err != nil {
				t.Errorf("Search for %s failed: %s", tc.in, err)
			}

			verify.Equal(t, out, tc.out, "Search for   "+tc.in)
		})
	}
}

func TestIndexWalk(t *testing.T) {
	idx, cleanFn := setupIdx(t)
	defer cleanFn()

	keys := populateIdx(t, idx)

	out := []string{}
	if err := idx.Walk(func(key string) error {
		out = append(out, key)
		return nil
	}); err != nil {
		t.Fatalf("Walking through index failed: %v", err)
	}

	verify.EqualSliceWithoutOrder(t, out, keys, "Walk through index")
}

func BenchmarkIndexCreationWithDefaultAnalyzer(b *testing.B) {
	idx, cleanFn := setupIdx(b)
	defer cleanFn()

	for n := 0; n < b.N; n++ {
		_ = populateIdx(b, idx)
	}

}

func BenchmarkIndexCreateWithfrAnalyzer(b *testing.B) {
	idx, cleanFn := setupIdx(b)
	defer cleanFn()

	idx.Mapping.DefaultAnalyzer = fr.AnalyzerName

	for n := 0; n < b.N; n++ {
		_ = populateIdx(b, idx)
	}
}
