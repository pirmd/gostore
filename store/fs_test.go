package store

import (
	"os"
	"strings"
	"testing"

	"github.com/pirmd/verify"
)

var (
	tstCases = []string{
		"file1.txt",
		"file2.txt",
		"file3.txt",
		"file4.txt",
	}

	validFn = func(s string) bool {
		return !strings.Contains(s, "_toFilter")
	}
)

func TestFSPut(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is availbale)

	for _, k := range tstCases {
		if err := fs.Put(NewRecord(k, nil), verify.IOReader("")); err != nil {
			t.Fatalf("Fail to add %v: %v", k, err)
		}
	}

	//This works because populate does not create subfolders
	tstDir.ShouldHaveContent(tstCases, "fs Put did not work as expected")
}

func TestFSWalk(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is availbale

	tstDir.Populate(tstCases)

	out := []string{}
	if err := fs.Walk(func(key string) error {
		out = append(out, key)
		return nil
	}); err != nil {
		t.Fatalf("Walking through fs failed: %v", err)
	}

	verify.EqualSliceWithoutOrder(t, out, tstCases, "Walk through fs does not work")
}

func TestFSDelete(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is availbale

	tstDir.Populate(tstCases)

	t.Run("Delete record", func(t *testing.T) {
		for _, tc := range tstCases {
			if err := fs.Delete(tc); err != nil {
				t.Errorf("Cannot delete '%s'", tc)
			}
			tstDir.ShouldNotHaveFile(tc, "Storefs cannot delete files")
		}
		tstDir.ShouldHaveContent([]string{}, "Storefs cannot delete files")
	})

	t.Run("Delete inexitant record", func(t *testing.T) {
		if err := fs.Delete("non_existing_record"); !os.IsNotExist(err) {
			t.Errorf("Should fail when deleting non existing record (err: %s)", err)
		}
	})

}

func TestFSForbiddenPath(t *testing.T) {
	tstDir := verify.NewTestField(t)
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is availbale

	tstCasesUnauthorized := []string{"folder_toFilter", "folder_toFilter/file.txt", "folder/file_toFilter.txt"}

	t.Run("Test Put()", func(t *testing.T) {
		for _, tc := range tstCasesUnauthorized {
			if err := fs.Put(NewRecord(tc, nil), verify.IOReader("")); err != os.ErrPermission {
				t.Errorf("Succeed to create '%v' that is forbidden", tc)
			}
		}
	})

	tstDir.Populate(tstCases)
	tstDir.Populate(tstCasesUnauthorized)

	t.Run("Test Get()", func(t *testing.T) {
		for _, tc := range tstCases {
			if _, err := fs.Get(tc); err != nil {
				t.Errorf("Cannot retrieve '%s'", tc)
			}
		}

		for _, tc := range tstCasesUnauthorized {
			if _, err := fs.Get(tc); err != os.ErrPermission {
				t.Errorf("Can retrieve '%s'", tc)
			}
		}
	})

	t.Run("Test Exists()", func(t *testing.T) {
		for _, tc := range tstCases {
			exists, err := fs.Exists(tc)
			if err != nil {
				t.Fatalf("Fail to check existence of '%s': %v", tc, err)
			}
			if !exists {
				t.Errorf("'%s' does not exists", tc)
			}
		}

		for _, tc := range tstCasesUnauthorized {
			_, err := fs.Exists(tc)
			if err != os.ErrPermission {
				t.Errorf("Acess to '%s' is not forbidden", tc)
			}
			if err != nil && err != os.ErrPermission {
				t.Errorf("Fail to check existence of '%s': %v", tc, err)
			}
		}
	})

	t.Run("Test Move()", func(t *testing.T) {
		for _, tc := range tstCases {
			if err := fs.Move(tc, NewRecord(tc+"_toFilter", nil)); err == nil {
				t.Errorf("Can move '%s' to '%s'", tc, tc+"_toFilter")
			}
		}

		for _, tc := range tstCasesUnauthorized {
			if err := fs.Move(tc, NewRecord(tc+"_moved", nil)); err == nil {
				t.Errorf("Can move '%s'", tc)
			}
		}

		for _, tc := range tstCases {
			if err := fs.Move(tc, NewRecord("moved/"+tc, nil)); err != nil {
				t.Errorf("Cannot move '%s' to '%s': %v", tc, "moved/"+tc, err)
			}
			tstDir.ShouldHaveFile("moved/"+tc, "storefs cannot move files")
			tstDir.ShouldNotHaveFile(tc, "storefs cannot move files")
		}
	})

	t.Run("Test Delete()", func(t *testing.T) {
		for _, tc := range tstCases {
			if err := fs.Delete("moved/" + tc); err != nil {
				t.Errorf("Cannot delete '%s'", "moved/"+tc)
			}
			tstDir.ShouldNotHaveFile("moved/"+tc, "storefs cannot delete files")
		}

		for _, tc := range tstCasesUnauthorized {
			if err := fs.Delete(tc); err == nil {
				t.Errorf("Can delete '%s'", tc)
			}
			tstDir.ShouldHaveFile(tc, "storefs can delete unauthorized files")
		}
	})
}
