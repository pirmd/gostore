package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/pirmd/clapp"
	"github.com/pirmd/verify"

	"github.com/pirmd/gostore/store"
)

const (
	testdataPath = "./testdata"
)

type testGostore struct {
	*Gostore
	*verify.TestFolder
}

func (gs *testGostore) Close() {
	gs.Gostore.Close()
	gs.TestFolder.Clean()
}

func newTestGostore(tb testing.TB, cfg *Config) *testGostore {
	tstPathName := strings.Replace(tb.Name(), string(os.PathSeparator), "_", -1)
	tstDir, err := verify.NewTestFolder(tstPathName)
	if err != nil {
		tb.Fatalf("Failed to create test folder: %v", err)
	}

	store.UseFrozenTimeStamps()

	cfg.UI.Auto = true
	cfg.Store.Path = tstDir.Root

	gs, err := openGostore(cfg)
	if err != nil {
		tb.Fatalf("cannot generate gostore from config: %s", err)
	}
	gs.log = verify.NewLogger(tb)

	return &testGostore{gs, tstDir}
}

func TestGostoreWithDefaultConfig(t *testing.T) {
	cfg := newConfig()

	httpmock := verify.StartMockHTTPResponse()
	defer httpmock.Stop()

	for _, style := range cfg.UI.ListStyles() {
		t.Run(style+"Fmt", func(t *testing.T) {
			cfg.UI.OutputFormat = style

			gs := newTestGostore(t, cfg)
			defer gs.Close()

			testImport(t, gs)
			testList(t, gs)
			testListAll(t, gs)
			testSearch(t, gs)
			testDelete(t, gs)
		})
	}
}

func TestGostoreWithConfigExample(t *testing.T) {
	//TODO: Directly reuse cmd.go code to load configuration file (for instance
	//if add ExpandEnv we need to do it twice and create mismatch)
	cfg := newConfig()
	appCfg := &clapp.Config{
		Unmarshaller: yaml.Unmarshal,
		Files:        []*clapp.ConfigFile{{Name: "config.example.yaml"}},
		Var:          cfg,
	}
	if err := appCfg.Load(); err != nil {
		t.Fatalf("Fail to read 'config.example.yaml': %v", err)
	}

	httpmock := verify.StartMockHTTPResponse()
	defer httpmock.Stop()

	for _, style := range cfg.UI.ListStyles() {
		t.Run(style+"Fmt", func(t *testing.T) {
			cfg.UI.OutputFormat = style

			gs := newTestGostore(t, cfg)
			defer gs.Close()

			testImport(t, gs)
			testList(t, gs)
			testListAll(t, gs)
			testSearch(t, gs)
			testDelete(t, gs)
		})
	}
}

func testImport(t *testing.T, gs *testGostore) {
	testCases, err := filepath.Glob(filepath.Join(testdataPath, "*.epub"))
	if err != nil {
		t.Fatalf("cannot read test data in %s:%v", testdataPath, err)
	}

	t.Run("ImportEpubs", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.Import(testCases); err != nil {
			t.Errorf("Fail to import epub '%s': %v", testCases, err)
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("Import output is not as expected.\n%v", failure)
		}

		if gs.store.IsDirty() {
			t.Errorf("Collection is inconsistent")
		}
	})

	t.Run("ImportTwiceEpubs", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.Import(testCases); err == nil {
			t.Errorf("Importing '%s' a second time worked but should not", testCases)
		}

		if failure := verify.EqualStdoutString(stdout, ""); failure != nil {
			t.Errorf("Importing an existing record should not print anything:\n%v", failure)
		}
	})

	t.Run("ImportNonExistingEpub", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.Import([]string{"non_existing.epub"}); err == nil {
			t.Errorf("Importing non existing epub worked but should not")
		}

		if failure := verify.EqualStdoutString(stdout, ""); failure != nil {
			t.Errorf("Importing non existing epub should not print anything:\n%v", failure)
		}
	})
}

func testList(t *testing.T, gs *testGostore) {
	allepubs, err := gs.ListWithExt(".epub")
	if err != nil {
		t.Fatalf("Fail to list epub: %v", err)
	}

	t.Run("ListEpubs", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.Gostore.List(allepubs); err != nil {
			t.Errorf("List failed: %v", err)
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("List output is not as expected:\n%v", failure)
		}
	})

	t.Run("ListNonExisting", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.Gostore.List([]string{"non existing record"}); err == nil {
			t.Errorf("Getting info for non existing record does no fail")
		}

		if failure := verify.EqualStdoutString(stdout, ""); failure != nil {
			t.Errorf("List output is not as expected:\n%v", failure)
		}
	})
}

func testListAll(t *testing.T, gs *testGostore) {
	t.Run("ListAll", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.ListAll(); err != nil {
			t.Fatalf("ListAll: fail to list epubs from collection: %v", err)
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("ListAll output is not as expected:\n%v", failure)
		}
	})
}

func testSearch(t *testing.T, gs *testGostore) {
	t.Run("SearchAll", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.Search("*"); err != nil {
			t.Fatalf("SearchAll: fail to list epubs from collection: %v", err)
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("SearchAll output is not as expected:\n%s", failure)
		}
	})

	t.Run("Search", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.Search("Adventures"); err != nil {
			t.Fatalf("Fail to search the collection: %v", err)
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("Search output is not as expected:\n%v", failure)
		}
	})

	//TODO(pirmd): add additional search pattern using date and serie number
}

func testDelete(t *testing.T, gs *testGostore) {
	allepubs, err := gs.ListWithExt(".epub")
	if err != nil {
		t.Fatalf("Fail to list epub: %v", err)
	}

	t.Run("DeleteEpubs", func(t *testing.T) {
		if err := gs.Delete([]string{allepubs[0]}); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		if err := gs.ListAll(); err != nil {
			t.Fatalf("ListAll after delete failed: cannot list epubs from collection: %v", err)
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("Delete output is not as expected:\n%v", failure)
		}
	})

	t.Run("DeleteNonExisting", func(t *testing.T) {
		if err := gs.Delete([]string{"non_existing.epub"}); err == nil {
			t.Errorf("deleting non existing record does no fail")
		}
	})
}
