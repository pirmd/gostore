package vfs

import (
	"github.com/pirmd/verify"
	"testing"
)

func TestCreate(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := NewOsfs()

	tc := "file.txt"
	if _, err := fs.Create(tstDir.Fullpath(tc)); err != nil {
		t.Errorf("Fail to create file '%s': %s", tc, err)
	}

	tstDir.ShouldHaveFile(tc, "Fail to create file")

	if _, err := fs.Open(tstDir.Fullpath(tc)); err != nil {
		t.Errorf("Fail to open file '%s': %s", tc, err)
	}
}

func TestMkdirAll(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	if err := fs.MkdirAll(tstDir.Fullpath("folder/subfolder"), 0777); err != nil {
		t.Errorf("Fail to create folder '%s': %s", "folder/subfolder", err)
	}

	if err := fs.MkdirAll(tstDir.Fullpath("folder/subfolder/subfolder"), 0777); err != nil {
		t.Errorf("Fail to create folder '%s': %s", "folder/subfolder/subfolder", err)
	}

	tstDir.ShouldHaveFile("folder/subfolder/subfolder", "cannot create a folder with its parents")
}

func TestRemoveAll(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	if err := fs.RemoveAll(tstDir.Fullpath("folder")); err != nil {
		t.Errorf("Failed to remove folder '%s': %s", "folder", err)
	}

	tstDir.ShouldHaveContent([]string{"file.txt"}, "Fail to remove all folders")
}
