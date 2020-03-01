package vfs

import (
	"testing"

	"github.com/pirmd/verify"
)

func TestWalk(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	ls, err := ListFs(fs, tstDir.Root)
	if err != nil {
		t.Fatalf("fail to list files in %s: %s", tstDir.Root, err)
	}

	if failure := tstDir.ShouldHaveContent(ls); failure != nil {
		t.Errorf("Fail to walk in a folder:\n%v", failure)
	}
}
