package store

import (
	"github.com/pirmd/verify"
	"path/filepath"
	"testing"
)

func setup_db(tb testing.TB) (*storedb, func()) {
	tstDir := verify.NewTestField(tb)

	db := newDB(filepath.Join(tstDir.Root, "test.db"))
	if err := db.Open(); err != nil {
		tstDir.Clean()
		tb.Fatalf("Fail to create testing Store database: %s", err)
	}

	return db, func() {
		db.Close()
		tstDir.Clean()
	}
}

func populate_db(tb testing.TB, db *storedb) (keys []string) {
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
	db, cleanFn := setup_db(t)
	defer cleanFn()

	keys := populate_db(t, db)

	out := []string{}
	if err := db.Walk(func(key string) error {
		out = append(out, key)
		return nil
	}); err != nil {
		t.Fatalf("Walking through db failed: %v", err)
	}

	verify.EqualSliceWithoutOrder(t, out, keys, "Walk through db")
}
