package vfs

import (
	"os"
	"strings"
	"testing"

	"github.com/pirmd/verify"
)

var (
	tstCasesUnauthorized = []string{
		"file_toFilter.txt",
		"folder/subfolder_toFilter",
		"folder_toFilter",
		"folder_toFilter/file.txt",
	}

	validFn = func(s string) bool {
		return !strings.Contains(s, "_toFilter")
	}
)

func TestFilterfsRead(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	fs := NewFilterfs(validFn, NewOsfs())

	t.Run("OpenFile", func(t *testing.T) {
		for _, tc := range tstCases {
			if _, err := fs.OpenFile(tstDir.Fullpath(tc), os.O_RDONLY, 0777); err != nil {
				t.Errorf("Fail to open '%s': %s", tc, err)
			}
		}
	})

	t.Run("OpenFile unauthorized", func(t *testing.T) {
		for _, tc := range tstCasesUnauthorized {
			if _, err := fs.OpenFile(tstDir.Fullpath(tc), os.O_RDONLY, 0777); err != os.ErrPermission {
				t.Errorf("Succeed to read unauthorized file '%s': %s", tc, err)
			}
		}
	})
}

func TestFilterfsPopulateAndWalk(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := NewFilterfs(validFn, NewOsfs())

	t.Run("Populate non filtered files", func(t *testing.T) {
		if err := PopulateFs(fs, tstDir.Root, tstCases); err != nil {
			t.Fatal(err)
		}
		if failure := tstDir.ShouldHaveContent(tstCases); failure != nil {
			t.Errorf("FilterFs cannot create tree of files and folders:\n%v", failure)
		}
	})

	t.Run("Populate filtered files", func(t *testing.T) {
		for _, tc := range tstCasesUnauthorized {
			if _, err := fs.Create(tstDir.Fullpath(tc)); err != os.ErrPermission {
				t.Errorf("Succeed to create an unauthorized file '%s'", tc)
			}
		}
		if failure := tstDir.ShouldHaveContent(tstCases); failure != nil {
			t.Errorf("FilterFs cannot create tree of files and folders:\n%v", failure)
		}
	})

	t.Run("Walk with non filtered files", func(t *testing.T) {
		tstDir.Populate(tstCasesUnauthorized)

		ls, err := ListFs(fs, tstDir.Root)
		if err != nil {
			t.Fatalf("fail to list files in %s: %s", tstDir.Root, err)
		}

		if failure := verify.EqualSliceWithoutOrder(ls, tstCases); failure != nil {
			t.Errorf("Filterfs cannot walk into folder with unauthorized file names:\n%v", failure)
		}
	})
}

func TestFilterfsRemove(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	fs := NewFilterfs(validFn, NewOsfs())

	t.Run("Remove file", func(t *testing.T) {
		tc := tstCases[4]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove file %s: %s", tc, err)
		}

		if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
			t.Errorf("Filterfs cannot remove file:\n%v", failure)
		}
	})

	t.Run("Remove unauthorized file", func(t *testing.T) {
		tc := tstCasesUnauthorized[3]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != os.ErrPermission {
			t.Errorf("Succeed to remove unauthorized file %s", tc)
		}
		if failure := tstDir.ShouldHaveFile(tc); failure != nil {
			t.Errorf("Filterfs does remove unauthorized file:\n%v", failure)
		}
	})

	t.Run("Remove empty folder", func(t *testing.T) {
		tc := tstCases[3] //Previous test has removed file inside this folder
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove empty folder %s: %s", tc, err)
		}
		if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
			t.Errorf("Filterfs cannot remove empty folder:\n%v", failure)
		}
	})

	t.Run("Remove unauthorized folder", func(t *testing.T) {
		tc := tstCasesUnauthorized[0]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != os.ErrPermission {
			t.Errorf("Succeed to remove unauthorized file %s", tc)
		}
		if failure := tstDir.ShouldHaveFile(tc); failure != nil {
			t.Errorf("Succeed to remove unauthorized file:\n%v", failure)
		}
	})

	t.Run("Remove non empty folder", func(t *testing.T) {
		tc := tstCases[1]
		if err := fs.Remove(tstDir.Fullpath(tc)); err == nil {
			t.Errorf("Succeed to remove non empty folder %s", tc)
		}
		if failure := tstDir.ShouldHaveFile(tc); failure != nil {
			t.Errorf("Filterfs does remove non empty folder:\n%v", failure)
		}
	})
}

func TestFilterfsRemoveAll(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	fs := NewFilterfs(validFn, NewOsfs())

	tc := "folder"
	if err := fs.RemoveAll(tstDir.Fullpath(tc)); err == nil {
		t.Errorf("Succeed to remove folders %s with unauthorized files: %s", tc, err)
	}

	if failure := tstDir.ShouldHaveContent(
		[]string{
			"file.txt",
			"file_toFilter.txt",
			"folder",
			"folder/subfolder_toFilter",
			"folder_toFilter",
			"folder_toFilter/file.txt",
		}); failure != nil {
		t.Errorf("Filterfs cannot remove recursively through a folder:\n%v", failure)
	}
}

func TestFilterfsRename(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	fs := NewFilterfs(validFn, NewOsfs())

	t.Run("Rename file", func(t *testing.T) {
		tc := tstCases[0]
		if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed")); err != nil {
			t.Errorf("Failed to rename %s: %s", tc, err)
		}

		if failure := tstDir.ShouldHaveFile(tc + "_renamed"); failure != nil {
			t.Errorf("Fileterfs failed to rename file:\n%v", failure)
		}
		if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
			t.Errorf("Filterfs failed to rename file:\n%v", failure)
		}
	})

	t.Run("Rename filtered file", func(t *testing.T) {
		tc := tstCasesUnauthorized[0]
		if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed")); err != os.ErrPermission {
			t.Errorf("Succeed to rename unauthorized file %s", tc)
		}

		if failure := tstDir.ShouldHaveFile(tc); failure != nil {
			t.Errorf("Fileterfs can rename unauthorized file:\n%v", failure)
		}
		if failure := tstDir.ShouldNotHaveFile(tc + "_renamed"); failure != nil {
			t.Errorf("Fileterfs can rename unauthorized file:\n%v", failure)
		}
	})

	t.Run("Rename file to unauthorized file", func(t *testing.T) {
		tc := tstCases[1]
		if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed_toFilter")); err != os.ErrPermission {
			t.Errorf("Succeed to rename %s to %s", tc, tc+"_toFilter")
		}

		if failure := tstDir.ShouldHaveFile(tc); failure != nil {
			t.Errorf("Filterfs can rename to unauthorized file:\n%v", failure)
		}
		if failure := tstDir.ShouldNotHaveFile(tc + "_renamed_toFilter"); failure != nil {
			t.Errorf("Filterfs can rename to unauthorized file:\n%v", failure)
		}
	})
}
