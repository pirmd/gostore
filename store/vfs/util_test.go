package vfs

import (
	"testing"

	"github.com/pirmd/verify"
)

func TestExists(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	tstList, err := tstDir.List()
	if err != nil {
		t.Fatalf("Fail to list test folder content: %v", err)
	}
	for _, f := range tstList {
		exists, err := fs.Exists(tstDir.Fullpath(f))
		if err != nil {
			t.Errorf("Failed to test existence of '%s': %s", f, err)
		}

		if !exists {
			t.Errorf("'%s' is seen as non existent", f)
		}
	}

	tstList, err = tstDir.List()
	if err != nil {
		t.Fatalf("Fail to list test folder content: %v", err)
	}
	for _, f := range tstList {
		exists, err := fs.Exists(tstDir.Fullpath(f) + "_nonexistent")
		if err != nil {
			t.Errorf("Failed to test existence of '%s': %s", f, err)
		}

		if exists {
			if failure := tstDir.ShouldNotHaveFile(f); failure != nil {
				t.Errorf("'%s' is seen as existent", f)
			}
		}
	}
}
