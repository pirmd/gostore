package vfs

import (
	"github.com/pirmd/verify"
	"os"
	"testing"
)

func TestOsfsRead(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	t.Run("OpenFile", func(t *testing.T) {
		for _, tc := range tstCases {
			if _, err := fs.OpenFile(tstDir.Fullpath(tc), os.O_RDONLY, 0777); err != nil {
				t.Errorf("Fail to open '%s': %s", tc, err)
			}
		}
	})
}

func TestOsfsPopulateAndWalk(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := NewOsfs()

	if err := PopulateFs(fs, tstDir.Root, tstCases); err != nil {
		t.Fatal(err)
	}

	tstDir.ShouldHaveContent(tstCases, "Fail to create tree")
}

func TestOsfsRemove(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	t.Run("Remove file", func(t *testing.T) {
		tc := tstCases[4]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove file %s: %s", tc, err)
		}

		tstDir.ShouldNotHaveFile(tc, "Failed to remove file")
	})

	t.Run("Remove empty folder", func(t *testing.T) {
		tc := tstCases[3] //Previous test has removed file inside this folder
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove empty folder %s: %s", tc, err)
		}
		tstDir.ShouldNotHaveFile(tc, "Failed to remove empty folder")
	})

	t.Run("Remove non empty folder", func(t *testing.T) {
		tc := tstCases[1]
		if err := fs.Remove(tstDir.Fullpath(tc)); err == nil {
			t.Errorf("Succeed to remove non empty folder %s", tc)
		}
		tstDir.ShouldHaveFile(tc, "Succeed to remove non empty folder")
	})
}

func TestOsfsRename(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	tc := tstCases[0]
	if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed")); err != nil {
		t.Errorf("Failed to rename %s: %s", tc, err)
	}

	tstDir.ShouldHaveFile(tc+"_renamed", "Failed to rename file")
	tstDir.ShouldNotHaveFile(tc, "Failed to rename file")
}
