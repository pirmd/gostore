package vfs

import (
	"os"
	"testing"

	"github.com/pirmd/verify"
)

func TestOsfsRead(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
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
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := NewOsfs()

	if err := PopulateFs(fs, tstDir.Root, tstCases); err != nil {
		t.Fatal(err)
	}

	if failure := tstDir.ShouldHaveContent(tstCases); failure != nil {
		t.Errorf("Fail to create tree:\n%v", failure)
	}
}

func TestOsfsRemove(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	t.Run("Remove file", func(t *testing.T) {
		tc := tstCases[4]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove file %s: %s", tc, err)
		}

		if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
			t.Errorf("Failed to remove file:\n%v", failure)
		}
	})

	t.Run("Remove empty folder", func(t *testing.T) {
		tc := tstCases[3] //Previous test has removed file inside this folder
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove empty folder %s: %s", tc, err)
		}
		if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
			t.Errorf("Failed to remove empty folder:\n%v", failure)
		}
	})

	t.Run("Remove non empty folder", func(t *testing.T) {
		tc := tstCases[1]
		if err := fs.Remove(tstDir.Fullpath(tc)); err == nil {
			t.Errorf("Succeed to remove non empty folder %s", tc)
		}
		if failure := tstDir.ShouldHaveFile(tc); failure != nil {
			t.Errorf("Succeed to remove non empty folder:\n%v", failure)
		}
	})
}

func TestOsfsRename(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	tc := tstCases[0]
	if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed")); err != nil {
		t.Errorf("Failed to rename %s: %s", tc, err)
	}

	if failure := tstDir.ShouldHaveFile(tc + "_renamed"); failure != nil {
		t.Errorf("Failed to rename file:\n%v", failure)
	}
	if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
		t.Errorf("Failed to rename file:\n%v", failure)
	}
}
