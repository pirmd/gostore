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
	tstDir := verify.NewTestField(t)
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
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := NewFilterfs(validFn, NewOsfs())

	t.Run("Populate non filtered files", func(t *testing.T) {
		if err := PopulateFs(fs, tstDir.Root, tstCases); err != nil {
			t.Fatal(err)
		}
		tstDir.ShouldHaveContent(tstCases, "FilterFs cannot create tree of files and folders")
	})

	t.Run("Populate filtered files", func(t *testing.T) {
		for _, tc := range tstCasesUnauthorized {
			if _, err := fs.Create(tstDir.Fullpath(tc)); err != os.ErrPermission {
				t.Errorf("Succeed to create an unauthorized file '%s'", tc)
			}
		}
		tstDir.ShouldHaveContent(tstCases, "FilterFs cannot create tree of files and folders")
	})

	t.Run("Walk with non filtered files", func(t *testing.T) {
		tstDir.Populate(tstCasesUnauthorized)

		ls, err := ListFs(fs, tstDir.Root)
		if err != nil {
			t.Fatalf("fail to list files in %s: %s", tstDir.Root, err)
		}

		verify.EqualSliceWithoutOrder(t, ls, tstCases, "Filterfs cannot walk into folder with unauthorized file names")
	})
}

func TestFilterfsRemove(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	fs := NewFilterfs(validFn, NewOsfs())

	t.Run("Remove file", func(t *testing.T) {
		tc := tstCases[4]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove file %s: %s", tc, err)
		}

		tstDir.ShouldNotHaveFile(tc, "Filterfs cannot remove file")
	})

	t.Run("Remove unauthorized file", func(t *testing.T) {
		tc := tstCasesUnauthorized[3]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != os.ErrPermission {
			t.Errorf("Succeed to remove unauthorized file %s", tc)
		}
		tstDir.ShouldHaveFile(tc, "Filterfs does remove unauthorized file")
	})

	t.Run("Remove empty folder", func(t *testing.T) {
		tc := tstCases[3] //Previous test has removed file inside this folder
		if err := fs.Remove(tstDir.Fullpath(tc)); err != nil {
			t.Errorf("Failed to remove empty folder %s: %s", tc, err)
		}
		tstDir.ShouldNotHaveFile(tc, "Filterfs cannot remove empty folder")
	})

	t.Run("Remove unauthorized folder", func(t *testing.T) {
		tc := tstCasesUnauthorized[0]
		if err := fs.Remove(tstDir.Fullpath(tc)); err != os.ErrPermission {
			t.Errorf("Succeed to remove unauthorized file %s", tc)
		}
		tstDir.ShouldHaveFile(tc, "Succeed to remove unauthorized file")
	})

	t.Run("Remove non empty folder", func(t *testing.T) {
		tc := tstCases[1]
		if err := fs.Remove(tstDir.Fullpath(tc)); err == nil {
			t.Errorf("Succeed to remove non empty folder %s", tc)
		}
		tstDir.ShouldHaveFile(tc, "Filterfs does remove non empty folder")
	})
}

func TestFilterfsRemoveAll(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	fs := NewFilterfs(validFn, NewOsfs())

	tc := "folder"
	if err := fs.RemoveAll(tstDir.Fullpath(tc)); err == nil {
		t.Errorf("Succeed to remove folders %s with unauthorized files: %s", tc, err)
	}

	tstDir.ShouldHaveContent([]string{"file.txt",
		"file_toFilter.txt",
		"folder",
		"folder/subfolder_toFilter",
		"folder_toFilter",
		"folder_toFilter/file.txt"},
		"Filterfs cannot remove recursively through a folder")
}

func TestFilterfsRename(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	fs := NewFilterfs(validFn, NewOsfs())

	t.Run("Rename file", func(t *testing.T) {
		tc := tstCases[0]
		if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed")); err != nil {
			t.Errorf("Failed to rename %s: %s", tc, err)
		}

		tstDir.ShouldHaveFile(tc+"_renamed", "Fileterfs failed to rename file")
		tstDir.ShouldNotHaveFile(tc, "Filterfs failed to rename file")
	})

	t.Run("Rename filtered file", func(t *testing.T) {
		tc := tstCasesUnauthorized[0]
		if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed")); err != os.ErrPermission {
			t.Errorf("Succeed to rename unauthorized file %s", tc)
		}

		tstDir.ShouldHaveFile(tc, "Fileterfs can rename unauthorized file")
		tstDir.ShouldNotHaveFile(tc+"_renamed", "Fileterfs can rename unauthorized file")
	})

	t.Run("Rename file to unauthorized file", func(t *testing.T) {
		tc := tstCases[1]
		if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed_toFilter")); err != os.ErrPermission {
			t.Errorf("Succeed to rename %s to %s", tc, tc+"_toFilter")
		}

		tstDir.ShouldHaveFile(tc, "Filterfs can rename to unauthorized file")
		tstDir.ShouldNotHaveFile(tc+"_renamed_toFilter", "Filterfs can rename to unauthorized file")
	})
}
