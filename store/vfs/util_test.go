package vfs

import (
	"testing"

	"github.com/pirmd/verify"
)

func TestExists(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	for _, f := range tstDir.List() {
		exists, err := fs.Exists(tstDir.Fullpath(f))
		if err != nil {
			t.Errorf("Failed to test existence of '%s': %s", f, err)
		}

		if !exists {
			t.Errorf("'%s' is seen as non existent", f)
		}
	}

	for _, f := range tstDir.List() {
		exists, err := fs.Exists(tstDir.Fullpath(f) + "_nonexistent")
		if err != nil {
			t.Errorf("Failed to test existence of '%s': %s", f, err)
		}

		if exists && !tstDir.Exists(f) {
			t.Errorf("'%s' is seen as existent", f)
		}
	}
}
