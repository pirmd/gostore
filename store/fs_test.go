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
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is available)

	for _, k := range tstCases {
		tr := NewRecord(k, nil)
		tr.SetFile(newMockROFile(""))

		if err := fs.Put(tr); err != nil {
			t.Fatalf("Fail to add %v: %v", k, err)
		}
	}

	//This works because populate does not create sub-folders
	if failure := tstDir.ShouldHaveContent(tstCases); failure != nil {
		t.Errorf("fs Put did not work as expected:\n%v", failure)
	}
}

func TestFSWalk(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is available

	tstDir.Populate(tstCases)

	out := []string{}
	if err := fs.Walk(func(key string) error {
		out = append(out, key)
		return nil
	}); err != nil {
		t.Fatalf("Walking through fs failed: %v", err)
	}

	if failure := verify.EqualSliceWithoutOrder(out, tstCases); failure != nil {
		t.Errorf("Walk through fs does not work:\n%v", failure)
	}
}

func TestFSSearchGlob(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is available

	tstDir.Populate(tstCases)

	out, err := fs.SearchGlob("file*.txt")
	if err != nil {
		t.Fatalf("Walking through fs failed: %v", err)
	}

	if failure := verify.EqualSliceWithoutOrder(out, tstCases); failure != nil {
		t.Errorf("Search through fs does not work:\n%v", failure)
	}
}

func TestFSDelete(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is available

	tstDir.Populate(tstCases)

	t.Run("Delete record", func(t *testing.T) {
		for _, tc := range tstCases {
			if err := fs.Delete(tc); err != nil {
				t.Errorf("Cannot delete '%s'", tc)
			}
			if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
				t.Errorf("Storefs cannot delete files:\n%v", failure)
			}
		}
		if failure := tstDir.ShouldHaveContent(nil); failure != nil {
			t.Errorf("storefs cannot delete files:\n%v", failure)
		}
	})

	t.Run("Delete inexitant record", func(t *testing.T) {
		if err := fs.Delete("non_existing_record"); !os.IsNotExist(err) {
			t.Errorf("Should fail when deleting non existing record (err: %s)", err)
		}
	})

}

func TestFSForbiddenPath(t *testing.T) {
	tstDir, err := verify.NewTestFolder(t.Name())
	if err != nil {
		t.Fatalf("Fail to create test folder: %v", err)
	}
	defer tstDir.Clean()

	fs := newFS(tstDir.Root, validFn)
	_ = fs.Open() //No need to check for error, (root path is tstDir and we know it is available

	tstCasesUnauthorized := []string{"folder_toFilter", "folder_toFilter/file.txt", "folder/file_toFilter.txt"}

	t.Run("Test Put()", func(t *testing.T) {
		for _, tc := range tstCasesUnauthorized {
			tr := NewRecord(tc, nil)
			tr.SetFile(newMockROFile(""))

			if err := fs.Put(tr); err != os.ErrPermission {
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
			if failure := tstDir.ShouldHaveFile("moved/" + tc); failure != nil {
				t.Errorf("storefs cannot move files:\n%v", failure)
			}
			if failure := tstDir.ShouldNotHaveFile(tc); failure != nil {
				t.Errorf("storefs cannot move files:\n%v", failure)
			}
		}
	})

	t.Run("Test Delete()", func(t *testing.T) {
		for _, tc := range tstCases {
			if err := fs.Delete("moved/" + tc); err != nil {
				t.Errorf("Cannot delete '%s'", "moved/"+tc)
			}
			if failure := tstDir.ShouldNotHaveFile("moved/" + tc); failure != nil {
				t.Errorf("storefs cannot delete files:\n%v", failure)
			}
		}

		for _, tc := range tstCasesUnauthorized {
			if err := fs.Delete(tc); err == nil {
				t.Errorf("Can delete '%s'", tc)
			}
			if failure := tstDir.ShouldHaveFile(tc); failure != nil {
				t.Errorf("storefs can delete unauthorized files:\n%v", failure)
			}
		}
	})
}
