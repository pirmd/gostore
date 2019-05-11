package store

import (
	"path/filepath"
	"testing"

	"github.com/blevesearch/bleve/analysis/lang/fr"
	"github.com/pirmd/verify"
)

func setup_store(tb testing.TB) (*Store, func()) {
	tstDir := verify.NewTestField(tb)

	s, err := Open(tstDir.Root)
	if err != nil {
		tstDir.Clean()
		tb.Fatalf("Fail to create testing Store: %s", err)
	}

	return s, func() {
		tstDir.Clean()
		s.Close()
	}
}

func populate_store(tb testing.TB, s *Store) (keys []string) {
	for _, td := range testData {
		r := NewRecord(buildKey(td), td)
		if err := s.Create(r, verify.IOReader("")); err != nil {
			tb.Fatalf("Fail to add %v: %s", td, err)
		}
		keys = append(keys, r.key)
	}
	return
}

func TestCreateAndRead(t *testing.T) {
	s, cleanFn := setup_store(t)
	defer cleanFn()

	keys := populate_store(t, s)

	t.Run("Can create a new record", func(t *testing.T) {
		for _, k := range keys {
			should_exist_in_store(t, s, k)
		}
	})

	t.Run("Can read a stored record", func(t *testing.T) {
		for i, tc := range testData {
			r, err := s.Read(keys[i])
			if err != nil {
				t.Errorf("Fail to retrieve '%s': %s", keys[i], err)
				continue
			}
			same_record_data(t, r, tc, "Mismatch between stored data and added data")
		}
	})

	t.Run("Cannot create an already existing record", func(t *testing.T) {
		for _, tc := range testData {
			if err := s.Create(NewRecord(buildKey(tc), tc), verify.IOReader("")); err == nil {
				t.Errorf("Managed to add %v that already exists in the store", tc)
			}
		}
	})

	t.Run("Can create a partially existing", func(t *testing.T) {
		for _, i := range []int{2, 4, 7, 9} {
			if err := s.db.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from database: %v", keys[i], err)
			}

			if err := s.Create(NewRecord(keys[i], testData[i]), verify.IOReader("")); err != nil {
				t.Errorf("Fail to create %s:%v: %v", keys[i], testData[i], err)
			}
		}

		for _, k := range keys {
			should_exist_in_store(t, s, k)
		}
	})

}

func TestDelete(t *testing.T) {
	s, cleanFn := setup_store(t)
	defer cleanFn()

	keys := populate_store(t, s)

	testCases := []int{1, 3, 5, 11}
	for _, i := range testCases {
		if err := s.Delete(keys[i]); err != nil {
			t.Fatalf("Fail to delete '%v': %s", keys[i], err)
		}
	}

	for i, _ := range testCases {
		if isIntInList(i, testCases) {
			should_not_exist_in_store(t, s, keys[i])
		} else {
			should_exist_in_store(t, s, keys[i])
		}
	}
}

func TestUpdate(t *testing.T) {
	s, cleanFn := setup_store(t)
	defer cleanFn()

	keys := populate_store(t, s)

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

			same_record_data(t, r, td, "Mismatch between stored data and updated data")
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
			should_not_exist_in_store(t, s, keys[i])
			should_exist_in_store(t, s, newkeys[i])

			r, err := s.Read(newkeys[i])
			if err != nil {
				t.Errorf("Fail to retrieve test case '%s': %s", newkeys[i], err)
				continue
			}
			same_record_data(t, r, td, "Mismatch between stored data and added data")
		}
	})

	t.Run("Cannot update partially existing", func(t *testing.T) {
		for i, td := range testData {
			if err := s.idx.Delete(newkeys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from index: %v", newkeys[i], err)
			}

			if err := s.Update(newkeys[i], NewRecord(newkeys[i], td)); err == nil {
				t.Errorf("Succeed to update partially existing Record '%v'", newkeys[i])
			}
		}
	})
}

func TestSearch(t *testing.T) {
	s, cleanFn := setup_store(t)
	defer cleanFn()

	if err := UsingDefaultAnalyzer(fr.AnalyzerName)(s); err != nil {
		t.Fatalf("Cannot choose 'fr' as default index analyzer: %v", err)
	}

	keys := populate_store(t, s)

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

		verify.EqualSliceWithoutOrder(t, out.Key(), tc.out, "Search:   "+tc.in)
	}
}

func TestCheckAndRepair(t *testing.T) {
	s, cleanFn := setup_store(t)
	defer cleanFn()

	keys := populate_store(t, s)

	testCases := []int{0, 2, 4, 6, 8}
	testCasesAfterDelete := []int{1, 3, 5, 7}
	orphansExpected := []string{}

	t.Run("Can repair missing from index", func(t *testing.T) {
		for _, i := range testCases {
			if err := s.idx.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from index: %v", keys[i], err)
			}
		}

		orphans, err := s.CheckAndRepair()
		if err != nil {
			t.Fatalf("Check and repair failed: %v", err)
		}
		verify.EqualSliceWithoutOrder(t, orphans, orphansExpected, "Orphans files mismatched")

		for i, _ := range testData {
			should_exist_in_store(t, s, keys[i])
		}
	})

	t.Run("Can Repair missing files", func(t *testing.T) {
		for _, i := range testCases {
			if err := s.fs.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from file-system: %v", keys[i], err)
			}
		}

		orphans, err := s.CheckAndRepair()
		if err != nil {
			t.Fatalf("Check and repair failed: %v", err)
		}
		verify.EqualSliceWithoutOrder(t, orphans, orphansExpected, "Orphans files mismatched")

		for i, _ := range testData {
			if isIntInList(i, testCases) {
				should_not_exist_in_store(t, s, keys[i])
			} else {
				should_exist_in_store(t, s, keys[i])
			}
		}
	})

	t.Run("Can repair missing from db", func(t *testing.T) {
		for _, i := range testCasesAfterDelete {
			if err := s.db.Delete(keys[i]); err != nil {
				t.Fatalf("Fail to delete '%s' from index: %v", keys[i], err)
			}
			orphansExpected = append(orphansExpected, keys[i])
		}

		orphans, err := s.CheckAndRepair()
		if err != nil {
			t.Fatalf("Check and repair failed: %v", err)
		}
		verify.EqualSliceWithoutOrder(t, orphans, orphansExpected, "Orphans files mismatched")

		for i, _ := range testData {
			if isIntInList(i, testCasesAfterDelete) {
				exists, err := s.idx.Exists(keys[i])
				if err != nil {
					t.Fatalf("Cannot check existence of '%s': %v", keys[i], err)
				}
				if exists {
					t.Errorf("'%s' was not cleared from the index", keys[i])
				}
			} else if isIntInList(i, testCases) {
				should_not_exist_in_store(t, s, keys[i])
			} else {
				should_exist_in_store(t, s, keys[i])
			}
		}
	})
}

func TestRebuildIndex(t *testing.T) {
	s, cleanFn := setup_store(t)
	defer cleanFn()

	keys := populate_store(t, s)

	err := s.RebuildIndex()
	if err != nil {
		t.Fatalf("Rebuild index failed: %v", err)
	}

	for i, _ := range testData {
		should_exist_in_store(t, s, keys[i])
	}
}

func same_record_data(tb testing.TB, r *Record, m map[string]interface{}, message string) {
	verify.EqualAsJson(tb, r.Value(), m, message)
}

func should_exist_in_store(tb testing.TB, s *Store, key string) {
	exists, err := s.Exists(key)
	if err != nil {
		tb.Errorf("Cannot check existence of '%s': %v", key, err)
	}

	if !exists {
		tb.Errorf("%#v does not exist in the Store but should", key)
	}
}

func should_not_exist_in_store(tb testing.TB, s *Store, key string) {
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
