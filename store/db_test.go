package store

import (
	"path/filepath"
	"testing"

	"github.com/pirmd/verify"
)

func setupDb(tb testing.TB) (*storedb, func()) {
	tstDir, err := verify.NewTestFolder(tb.Name())
	if err != nil {
		tb.Fatalf("Fail to create test folder: %v", err)
	}

	db := newDB(filepath.Join(tstDir.Root, "test.db"))
	if err := db.Open(); err != nil {
		tstDir.Clean()
		tb.Fatalf("Fail to create testing Store database: %s", err)
	}

	return db, func() {
		tstDir.Clean()
		if err := db.Close(); err != nil {
			tb.Fatalf("Fail to properly close testing database: %s", err)
		}
	}
}

func populateDb(tb testing.TB, db *storedb) (keys []string) {
	for _, td := range testData {
		r := NewRecord(buildKey(td), td)

		if err := db.Put(r); err != nil {
			tb.Fatalf("Fail to add %v", td)
		}

		keys = append(keys, r.key)
	}
	return
}

func TestDBWalk(t *testing.T) {
	db, cleanFn := setupDb(t)
	defer cleanFn()

	keys := populateDb(t, db)

	out := []string{}
	if err := db.Walk(func(key string) error {
		out = append(out, key)
		return nil
	}); err != nil {
		t.Fatalf("Walking through db failed: %v", err)
	}

	if failure := verify.EqualSliceWithoutOrder(out, keys); failure != nil {
		t.Errorf("Walk through db failed:\n%v", failure)
	}
}
