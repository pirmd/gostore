package vfs

import (
	"testing"

	"github.com/pirmd/verify"
)

func TestWalk(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	ls, err := ListFs(fs, tstDir.Root)
	if err != nil {
		t.Fatalf("fail to list files in %s: %s", tstDir.Root, err)
	}

	tstDir.ShouldHaveContent(ls, "Fail to walk in a folder")
}
