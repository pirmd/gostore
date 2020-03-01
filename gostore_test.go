package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/pirmd/verify"

	"github.com/pirmd/gostore/store"
)

const (
	testdataPath = "./testdata"
)

var (
	debug = flag.Bool("test.debug", false, "print debug information from gostore during testing steps")
)

type testGostore struct {
	*Gostore
	*verify.TestFolder
}

func newTestGostore(tb testing.TB, cfg *Config) *testGostore {
	cfg.ShowLog = *debug
	cfg.UI.Auto = true

	tstDir, err := verify.NewTestFolder(tb.Name())
	if err != nil {
		tb.Fatalf("Failed to create test folder: %v", err)
	}

	if cfg.Store == nil {
		cfg.Store = &storeConfig{}
	}
	cfg.Store.Root = tstDir.Root

	gs, err := newGostore(cfg)
	if err != nil {
		tb.Fatalf("cannot generate gostore from config: %s", err)
	}

	if err := store.UsingFrozenTimeStamps()(gs.store); err != nil {
		tb.Fatalf("cannot force store to use frozen time-stamps for test duration: %s", err)
	}

	return &testGostore{gs, tstDir}
}

func TestGostoreWithDefaultConfig(t *testing.T) {
	cfg := newConfig()

	httpmock := verify.StartMockHTTPResponse()
	defer httpmock.Stop()

	gs := newTestGostore(t, cfg)
	defer gs.Clean()

	testImport(t, gs)
	testInfo(t, gs)
	testListAll(t, gs)
	testSearch(t, gs)
	testDelete(t, gs)
}

func TestGostoreWithConfigExample(t *testing.T) {
	b, err := ioutil.ReadFile("config.example.yaml")
	if err != nil {
		t.Fatalf("Fail to read 'config.example.yaml': %v", err)
	}

	bexpanded := []byte(os.ExpandEnv(string(b)))

	cfg := newConfig()
	if err := yaml.Unmarshal(bexpanded, cfg); err != nil {
		t.Fatalf("Fail to read 'config.example.yaml': %v", err)
	}

	httpmock := verify.StartMockHTTPResponse()
	defer httpmock.Stop()

	gs := newTestGostore(t, cfg)
	defer gs.Clean()

	testImport(t, gs)
	testInfo(t, gs)
	testListAll(t, gs)
	testSearch(t, gs)
	testDelete(t, gs)
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

		for _, tc := range testCases {
			if err := gs.Import(tc); err != nil {
				t.Errorf("Fail to import epub '%s': %v", tc, err)
			}
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("Import output is not as expected.\n%v", failure)
		}

		// TODO(pirmd): update store's api to get quicker information regarding store's consitenbcy state
		if err := gs.store.Open(); err != nil {
			t.Errorf("Collection is inconsistent: %v", err)
		}
		defer gs.store.Close()

		orphans, err := gs.store.CheckAndRepair()
		if err != nil {
			t.Errorf("Collection is inconsistent: %v", err)
		}
		if len(orphans) > 0 {
			t.Errorf("Collection is inconsistent: several orphans have been found: %v", orphans)
		}
	})

	t.Run("ImportTwiceEpubs", func(t *testing.T) {
		for _, tc := range testCases {
			if err := gs.Import(tc); err == nil {
				t.Errorf("Importing '%s' a second time worked but should not", tc)
			}
		}
	})

	t.Run("ImportNonExistingEpub", func(t *testing.T) {
		if err := gs.Import("non_existing.epub"); err == nil {
			t.Errorf("Importing non existing epub worked but should not")
		}
	})
}

func testInfo(t *testing.T, gs *testGostore) {
	allepubs, err := gs.ListWithExt(".epub")
	if err != nil {
		t.Fatalf("Fail to list epub: %v", err)
	}

	t.Run("InfoEpubs", func(t *testing.T) {
		stdout, err := verify.StartMockStdout()
		if err != nil {
			t.Fatalf("Fail to mock stdout: %v", err)
		}
		defer stdout.Stop()

		for _, epub := range allepubs {
			if err := gs.Info(epub, false); err != nil {
				t.Errorf("Getting Info failed: %v", err)
			}
		}

		if failure := verify.MatchStdoutGolden(t.Name(), stdout); failure != nil {
			t.Errorf("Info output is not as expected:\n%v", failure)
		}
	})

	t.Run("InfoNonExisting", func(t *testing.T) {
		if err := gs.Info("non existing record", false); err == nil {
			t.Errorf("Getting info for non existing record does no fail")
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

		if err := gs.Search("Alice"); err != nil {
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
		if err := gs.Delete(allepubs[0]); err != nil {
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
		if err := gs.Delete("non_existing.epub"); err == nil {
			t.Errorf("deleting non existing record does no fail")
		}
	})
}
