package vfs

import (
	"os"
	"strings"
	"testing"

	"github.com/pirmd/verify"
)

func TestReadonlyfsRead(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewReadonlyfs(NewOsfs())

	t.Run("OpenFile", func(t *testing.T) {
		for _, tc := range tstCases {
			if _, err := fs.OpenFile(tstDir.Fullpath(tc), os.O_RDONLY, 0777); err != nil {
				t.Errorf("Fail to open '%s': %s", tc, err)
			}
		}
	})

}

func TestReadonlyfsPopulateAndWalk(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := NewReadonlyfs(NewOsfs())

	t.Run("Populate should fail", func(t *testing.T) {
		if err := PopulateFs(fs, tstDir.Root, tstCases); err != os.ErrPermission {
            t.Errorf("Succeed to add file '%s': %s", tc, err)
		}
	})

    tstDir.Populate(tstCases)

    t.Run("OpenFile for writing should fail", func(t *testing.T) {
        for _, tc := range tstCases {
            if _, err := fs.OpenFile(tstDir.Fullpath(tc), os.O_WRONLY, 0777); err != os.ErrPermission {
                t.Errorf("Succeed to open for writing a file '%s': %s", tc, err)
            }
        }
    })

	t.Run("Walk", func(t *testing.T) {
		ls, err := ListFs(fs, tstDir.Root)
		if err != nil {
			t.Fatalf("fail to list files in %s: %s", tstDir.Root, err)
		}

		verify.EqualSliceWithoutOrder(t, ls, tstCases, "Readonlyfs cannot walk into folder")
	})
}

func TestReadonlyfsRemove(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewReadonlyfs(NewOsfs())

	t.Run("Remove file should fail", func(t *testing.T) {
		for _, tc := range tstCases {
            if err := fs.Remove(tstDir.Fullpath(tc)); err != os.ErrPermission {
                t.Errorf("Succeed to remove file %s: %s", tc, err)
            }
        }
	})
}

func TestReadonlyfsRename(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	tstDir.Populate(tstCases)

	fs := NewReadonlyfs(NewOsfs())

	t.Run("Rename file should fail", func(t *testing.T) {
		for _, tc := range tstCases {
            if err := fs.Rename(tstDir.Fullpath(tc), tstDir.Fullpath(tc+"_renamed")); err != os.ErrPermission {
                t.Errorf("Succeed to rename %s: %s", tc, err)
            }
        }
	})
}
