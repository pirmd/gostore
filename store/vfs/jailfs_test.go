package vfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pirmd/verify"
)

func TestJailfsPath(t *testing.T) {
	testCases := []struct {
		in  string
		out string
	}{
		{"", "test"},
		{"test1", filepath.Join("test", "test1")},
		{"./test1", filepath.Join("test", "test1")},
		{"/test1", filepath.Join("test", "test1")},
		{"../test1", filepath.Join("test", "test1")},
		{"../test/test1", filepath.Join("test", "test", "test1")},
	}

	fs := &jailfs{"test", NewOsfs()}
	for _, tc := range testCases {
		out := fs.realPath(tc.in)
		if out != tc.out {
			t.Errorf("Real path failed\nGet     : %s\nExpected: %s\n", out, tc.out)
		}
	}
}

func TestJailfsRead(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewJailfs(tstDir.Root, NewOsfs())

	t.Run("OpenFile", func(t *testing.T) {
		for _, tc := range tstCases {
			if _, err := fs.OpenFile(tc, os.O_RDONLY, 0777); err != nil {
				t.Errorf("Fail to open '%s': %s", tc, err)
			}
		}
	})
}

func TestJailfsPopulateAndWalk(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := NewJailfs(tstDir.Root, NewOsfs())

	if err := PopulateFs(fs, "", tstCases); err != nil {
		t.Fatal(err)
	}

	tstDir.ShouldHaveContent(tstCases, "Fail to create tree")
}

func TestJailfsRemove(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewJailfs(tstDir.Root, NewOsfs())

	t.Run("Remove file", func(t *testing.T) {
		tc := tstCases[4]
		if err := fs.Remove(tc); err != nil {
			t.Errorf("Failed to remove file %s: %s", tc, err)
		}
		tstDir.ShouldNotHaveFile(tc, "Failed to remove file")
	})

	t.Run("Remove empty folder", func(t *testing.T) {
		tc := tstCases[3] //Previous test has removed file inside this folder
		if err := fs.Remove(tc); err != nil {
			t.Errorf("Failed to remove empty folder %s: %s", tc, err)
		}
		tstDir.ShouldNotHaveFile(tc, "Failed to remove empty folder")
	})

	t.Run("Remove non empty folder", func(t *testing.T) {
		tc := tstCases[1]
		if err := fs.Remove(tc); err == nil {
			t.Errorf("Succeed to remove non empty folder %s", tc)
		}
		tstDir.ShouldHaveFile(tc, "Succeed to remove non empty folder")
	})

	t.Run("Remove root folder", func(t *testing.T) {
		if err := fs.RemoveAll(""); err != os.ErrPermission {
			t.Errorf("Succeed to remove root folder")
		}
	})
	tstDir.ShouldHaveContent([]string{""}, "Failed to remove folders")
}

func TestJailfsRename(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewJailfs(tstDir.Root, NewOsfs())

	t.Run("Rename file", func(t *testing.T) {
		tc := tstCases[0]
		if err := fs.Rename(tc, tc+"_renamed"); err != nil {
			t.Errorf("Failed to rename %s: %s", tc, err)
		}

		tstDir.ShouldHaveFile(tc+"_renamed", "Failed to rename file")
		tstDir.ShouldNotHaveFile(tc, "Failed to rename file")
	})

	t.Run("Rename root", func(t *testing.T) {
		if err := fs.Rename("", "root_renamed"); err != os.ErrPermission {
			t.Errorf("Succeed to rename root folder")
		}

		tstDir.ShouldHaveFile("", "Succeed to rename root folder")
		tstDir.ShouldNotHaveFile("root_renamed", "Succeed to rename root folder")
	})
}
