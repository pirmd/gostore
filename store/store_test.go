package store

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/blevesearch/bleve/analysis/lang/fr"
	"github.com/pirmd/verify"
)

func setupStore(tb testing.TB) (*Store, func()) {
	tstDir, err := verify.NewTestFolder(tb.Name())
	if err != nil {
		tb.Fatalf("Fail to create test folder: %v", err)
	}

	UseFrozenTimeStamps()
	s, err := New(tstDir.Root)
	if err != nil {
		tstDir.Clean()
		tb.Fatalf("Fail to create testing Store: %s", err)
	}

	if err := s.Open(); err != nil {
		tstDir.Clean()
		tb.Fatalf("Fail to open testing Store: %s", err)
	}

	return s, func() {
		tstDir.Clean()
		if err := s.Close(); err != nil {
			tb.Fatalf("Fail to properly close testing Store: %s", err)
		}
	}
}

func populateStore(tb testing.TB, s *Store) (keys []string) {
	for _, td := range testData {
		r := NewRecord(buildKey(td), td)
		if err := s.Create(r, verify.MockROFile("")); err != nil {
			tb.Fatalf("Fail to add %v: %s", td, err)
		}
		keys = append(keys, r.key)
	}
	return
}

func TestCreateAndRead(t *testing.T) {
	s, cleanFn := setupStore(t)
	defer cleanFn()

	keys := populateStore(t, s)

	t.Run("Can create a new record", func(t *testing.T) {
		for _, k := range keys {
			shouldExistInStore(t, s, k)
		}
	})

	t.Run("Can read a stored record", func(t *testing.T) {
		for i, tc := range testData {
			r, err := s.Read(keys[i])
			if err != nil {
				t.Errorf("Fail to retrieve '%s': %s", keys[i], err)
				continue
			}
			sameRecordData(t, r, tc, "Mismatch between stored data and added data")
		}
	})

	t.Run("Cannot create an already existing record", func(t *testing.T) {
		for _, tc := range testData {
			if err := s.Create(NewRecord(buildKey(tc), tc), verify.MockROFile("")); err == nil {
				t.Errorf("Managed to add %v that already exists in the store", tc)
			}
		}
	})

	t.Run("Can create a partially existing", func(t *testing.T) {
		for _, i := range []int{2, 4, 7, 9} {
			if err := s.db.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from database: %v", keys[i], err)
			}

			if err := s.Create(NewRecord(keys[i], testData[i]), verify.MockROFile("")); err != nil {
				t.Errorf("Fail to create %s:%v: %v", keys[i], testData[i], err)
			}
		}

		for _, k := range keys {
			shouldExistInStore(t, s, k)
		}
	})

}

func TestReadGlob(t *testing.T) {
	s, cleanFn := setupStore(t)
	defer cleanFn()

	_ = populateStore(t, s)

	testPattern := "Luc*/*.tst"
	var want []map[string]interface{}
	for _, i := range []int{0, 1} {
		want = append(want, testData[i])
	}

	got, err := s.ReadGlob(testPattern)
	if err != nil {
		t.Fatalf("Fail to retrieve '%s': %s", testPattern, err)
	}

	sameRecordsData(t, got, want, "Read pattern "+testPattern+" failed")
}

func TestDelete(t *testing.T) {
	s, cleanFn := setupStore(t)
	defer cleanFn()

	keys := populateStore(t, s)

	testCases := []int{1, 3, 5, 11}
	for _, i := range testCases {
		if err := s.Delete(keys[i]); err != nil {
			t.Fatalf("Fail to delete '%v': %s", keys[i], err)
		}
	}

	for i := range testData {
		if isIntInList(i, testCases) {
			shouldNotExistInStore(t, s, keys[i])
		} else {
			shouldExistInStore(t, s, keys[i])
		}
	}
}

func TestUpdate(t *testing.T) {
	s, cleanFn := setupStore(t)
	defer cleanFn()

	keys := populateStore(t, s)

	for _, td := range testData {
		td["updated"] = true
	}

	t.Run("Can update Record.value", func(t *testing.T) {
		for i, td := range testData {
			if err := s.Update(keys[i], NewRecord(keys[i], td)); err != nil {
				t.Fatalf("Fail to update: %v", err)
			}
		}

		for i, td := range testData {
			r, err := s.Read(keys[i])
			if err != nil {
				t.Errorf("Fail to retrieve test case '%s': %s", keys[i], err)
				continue
			}

			sameRecordData(t, r, td, "Mismatch between stored data and updated data")
		}
	})

	newkeys := make([]string, len(keys))
	t.Run("Can update Record.key", func(t *testing.T) {
		for i, td := range testData {
			newkeys[i] = filepath.Join("updated", keys[i])
			if err := s.Update(keys[i], NewRecord(newkeys[i], td)); err != nil {
				t.Fatalf("Fail to update '%#v': %v", keys[i], err)
			}
		}

		for i, td := range testData {
			shouldNotExistInStore(t, s, keys[i])
			shouldExistInStore(t, s, newkeys[i])

			r, err := s.Read(newkeys[i])
			if err != nil {
				t.Errorf("Fail to retrieve test case '%s': %s", newkeys[i], err)
				continue
			}
			sameRecordData(t, r, td, "Mismatch between stored data and added data")
		}
	})
}

func TestSearch(t *testing.T) {
	s, cleanFn := setupStore(t)
	defer cleanFn()

	if err := UsingDefaultAnalyzer(fr.AnalyzerName)(s); err != nil {
		t.Fatalf("Cannot choose 'fr' as default index analyzer: %v", err)
	}

	keys := populateStore(t, s)

	testCases := []struct {
		in  string
		out []string
	}{
		{"épée", []string{keys[5], keys[12]}},
		{"charles michel", []string{keys[12]}},
		{"Nettoyage", []string{keys[3]}},
		{"l'Épée Charles-Michel (de)", []string{keys[5], keys[12]}},
		{"miserable", []string{keys[2]}},
	}

	for _, tc := range testCases {
		out, err := s.Search(tc.in)
		if err != nil {
			t.Errorf("Search for %s failed: %s", tc.in, err)
		}

		if failure := verify.EqualSliceWithoutOrder(out.Key(), tc.out); failure != nil {
			t.Errorf("Search for %s failed:\n%v", tc.in, failure)
		}
	}
}

func TestCheckAndRepair(t *testing.T) {
	s, cleanFn := setupStore(t)
	defer cleanFn()

	keys := populateStore(t, s)

	if s.IsDirty() {
		t.Errorf("Store is dirty")
	}

	testCases := []int{0, 2, 4, 6, 8}
	testCasesAfterDelete := []int{1, 3, 5, 7}

	t.Run("Can repair missing from index", func(t *testing.T) {
		for _, i := range testCases {
			if err := s.idx.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from index: %v", keys[i], err)
			}
		}

		err := s.RepairIndex()
		if err != nil {
			t.Fatalf("Check and repair failed: %v", err)
		}
		if s.IsDirty() {
			t.Errorf("Store is dirty")
		}

		for i := range testData {
			shouldExistInStore(t, s, keys[i])
		}
	})

	t.Run("Can detect missing from db", func(t *testing.T) {
		orphansExpected := []string{}

		for _, i := range testCases {
			if err := s.db.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from file-system: %v", keys[i], err)
			}
			orphansExpected = append(orphansExpected, keys[i])
		}

		orphans, err := s.CheckOrphans()
		if err != nil {
			t.Fatalf("Check for orphans failed: %v", err)
		}
		if failure := verify.EqualSliceWithoutOrder(orphans, orphansExpected); failure != nil {
			t.Errorf("Orphans files mismatched:\n%v", failure)
		}
	})

	t.Run("Can repair missing files", func(t *testing.T) {
		ghostsExpected := []string{}
		for _, i := range testCasesAfterDelete {
			if err := s.fs.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from index: %v", keys[i], err)
			}
			ghostsExpected = append(ghostsExpected, keys[i])
		}

		ghosts, err := s.CheckGhosts()
		if err != nil {
			t.Fatalf("Check for ghosts: %v", err)
		}
		if failure := verify.EqualSliceWithoutOrder(ghosts, ghostsExpected); failure != nil {
			t.Errorf("Ghosts files mismatched:\n%v", failure)
		}
	})
}

func TestRebuildIndex(t *testing.T) {
	s, cleanFn := setupStore(t)
	defer cleanFn()

	keys := populateStore(t, s)

	err := s.RebuildIndex()
	if err != nil {
		t.Fatalf("Rebuild index failed: %v", err)
	}

	for i := range testData {
		shouldExistInStore(t, s, keys[i])
	}
}

func sameRecordData(tb testing.TB, r *Record, m map[string]interface{}, message string) {
	tb.Helper()

	rInJSON, err := json.Marshal(r.UserValue())
	if err != nil {
		tb.Fatalf("Failed to marshall to JSON: %v", err)
	}

	mInJSON, err := json.Marshal(m)
	if err != nil {
		tb.Fatalf("Failed to marshal to JSON: %v", err)
	}

	if failure := verify.Equal(string(rInJSON), string(mInJSON)); failure != nil {
		tb.Errorf("%s:\n%v", message, failure)
	}
}

func sameRecordsData(tb testing.TB, rec Records, maps []map[string]interface{}, message string) {
	tb.Helper()

	rInJSON := []string{}
	for _, r := range rec {
		j, err := json.Marshal(r.UserValue())
		if err != nil {
			tb.Fatalf("Failed to marshall to JSON: %v", err)
		}
		rInJSON = append(rInJSON, string(j))
	}

	mInJSON := []string{}
	for _, m := range maps {
		j, err := json.Marshal(m)
		if err != nil {
			tb.Fatalf("Failed to marshal to JSON: %v", err)
		}
		mInJSON = append(mInJSON, string(j))
	}

	if failure := verify.EqualSliceWithoutOrder(rInJSON, mInJSON); failure != nil {
		tb.Errorf("%s:\n%v", message, failure)
	}
}

func shouldExistInStore(tb testing.TB, s *Store, key string) {
	tb.Helper()

	exists, err := s.Exists(key)
	if err != nil {
		tb.Errorf("Cannot check existence of '%s': %v", key, err)
	}

	if !exists {
		tb.Errorf("%#v does not exist in the Store but should", key)
	}
}

func shouldNotExistInStore(tb testing.TB, s *Store, key string) {
	tb.Helper()

	exists, err := s.Exists(key)
	if err != nil {
		tb.Errorf("Cannot check existence of '%s': %v", key, err)
	}

	if exists {
		tb.Errorf("%#v does exist in the Store but should", key)
	}
}

func isIntInList(i int, list []int) bool {
	for _, l := range list {
		if i == l {
			return true
		}
	}
	return false
}
