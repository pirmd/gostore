package vfs

import (
	"testing"

	"github.com/pirmd/verify"
)

func TestCreate(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := NewOsfs()

	tc := "file.txt"
	if _, err := fs.Create(tstDir.Fullpath(tc)); err != nil {
		t.Errorf("Fail to create file '%s': %s", tc, err)
	}

	if failure := tstDir.ShouldHaveFile(tc); failure != nil {
		t.Errorf("Fail to create file:\n%v", failure)
	}

	if _, err := fs.Open(tstDir.Fullpath(tc)); err != nil {
		t.Errorf("Fail to open file '%s': %s", tc, err)
	}
}

func TestMkdirAll(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	if err := fs.MkdirAll(tstDir.Fullpath("folder/subfolder"), 0777); err != nil {
		t.Errorf("Fail to create folder '%s': %s", "folder/subfolder", err)
	}

	if err := fs.MkdirAll(tstDir.Fullpath("folder/subfolder/subfolder"), 0777); err != nil {
		t.Errorf("Fail to create folder '%s': %s", "folder/subfolder/subfolder", err)
	}

	if failure := tstDir.ShouldHaveFile("folder/subfolder/subfolder"); failure != nil {
		t.Errorf("Cannot create a folder with its parents:\n%v", failure)
	}
}

func TestRemoveAll(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewOsfs()

	if err := fs.RemoveAll(tstDir.Fullpath("folder")); err != nil {
		t.Errorf("Failed to remove folder '%s': %s", "folder", err)
	}

	if failure := tstDir.ShouldHaveContent([]string{"file.txt"}); failure != nil {
		t.Errorf("Fail to remove all folders:\n%v", failure)
	}
}
