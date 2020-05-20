package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui"
	"github.com/pirmd/gostore/ui/cli"
)

// Config represents the configuration for gostore
type Config struct {
	// ShowLog is a flag that governs if log information are to be shown
	ShowLog bool

	// ReadOnly is the flag to switch the store into read only operation mode
	ReadOnly bool

	// Store contains configuration for anything related to storage
	Store *store.Config

	// UI contains configuration for anything related to user interface
	UI *cli.Config

	// ImportModules lists of modules to be applied when importing a record
	ImportModules map[string]*rawYAMLConfig

	// UpdateModules lists of modules to be applied when updating a record
	UpdateModules map[string]*rawYAMLConfig
}

func (cfg *Config) expandEnv() {
	cfg.Store.Path = os.ExpandEnv(cfg.Store.Path)
	cfg.UI.EditorCmd = os.ExpandEnv(cfg.UI.EditorCmd)
	cfg.UI.MergerCmd = os.ExpandEnv(cfg.UI.MergerCmd)
}

func newConfig() *Config {
	return &Config{
		Store: store.NewConfig(),
		UI:    cli.NewConfig(),
	}
}

type rawYAMLConfig struct {
	unmarshal func(interface{}) error
}

func (cfg *rawYAMLConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	cfg.unmarshal = unmarshal
	return nil
}

func (cfg *rawYAMLConfig) Unmarshal(v interface{}) error {
	if cfg == nil {
		return nil
	}

	return cfg.unmarshal(v)
}

// Gostore represents the main collection manager.
type Gostore struct {
	log           *log.Logger
	pretend       bool
	store         *store.Store
	ui            ui.UserInterfacer
	importModules []modules.Module
	updateModules []modules.Module
}

func newGostore(cfg *Config) (*Gostore, error) {
	cfg.expandEnv()

	gs := &Gostore{
		log:     log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		pretend: cfg.ReadOnly,
	}

	if cfg.ShowLog {
		gs.log.SetOutput(os.Stderr)
		cfg.Store.Logger = gs.log
	}

	var err error
	if gs.store, err = store.NewFromConfig(cfg.Store); err != nil {
		return nil, err
	}

	if gs.ui, err = cli.NewFromConfig(cfg.UI); err != nil {
		return nil, err
	}

	for modName, modRawCfg := range cfg.ImportModules {
		m, err := modules.New(modName, modRawCfg, gs.log, gs.ui)
		if err != nil {
			return nil, err
		}

		gs.importModules = append(gs.importModules, m)
	}

	for modName, modRawCfg := range cfg.UpdateModules {
		m, err := modules.New(modName, modRawCfg, gs.log, gs.ui)
		if err != nil {
			return nil, fmt.Errorf("cannot create module '%s': %v", modName, err)
		}

		gs.updateModules = append(gs.updateModules, m)
	}

	return gs, nil
}

// Import adds new media into the collection
func (gs *Gostore) Import(mediaFiles []string) error {
	var newRecords store.Records

	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	//TODO: should be in store or in an util package?
	importErr := new(store.NonBlockingErrors)

	for _, path := range mediaFiles {
		gs.log.Printf("Adding media file %s to the collection", path)
		f, err := os.Open(path)
		if err != nil {
			importErr.Add(fmt.Errorf("failed to import %s: %s", path, err))
			continue
		}
		defer f.Close()

		mdataFromFile, err := media.GetMetadata(f)
		if err != nil {
			importErr.Add(fmt.Errorf("failed to import %s: %s", path, err))
			continue
		}
		r := store.NewRecord(filepath.Base(path), mdataFromFile)

		if err := modules.ProcessRecord(r, gs.importModules); err != nil {
			importErr.Add(fmt.Errorf("failed to import %s: %s", path, err))
			continue
		}

		if !gs.pretend {
			gs.log.Printf("Creating new record %v to the collection", r)
			if err := gs.store.Create(r, f); err != nil {
				importErr.Add(fmt.Errorf("failed to import %s: %s", path, err))
				continue
			}
		}

		newRecords = append(newRecords, r)
	}

	if len(newRecords) != 0 {
		gs.ui.PrettyPrint(newRecords.Flatted()...)
	}

	return importErr.Err()
}

// Info retrieves information about any collection's record.
// If fromFile flag is set, Info also displays difference with actual metadata
// stored in the media file.
func (gs *Gostore) Info(key string, fromFile bool) error {
	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	r, err := gs.store.Read(key)
	if err != nil {
		return err
	}

	if fromFile {
		f, err := gs.store.OpenRecord(key)
		if err != nil {
			return err
		}
		defer f.Close()

		mdata, err := media.GetMetadata(f)
		if err != nil {
			return err
		}

		gs.ui.PrettyDiff(r.Flatted(), mdata)
		return nil
	}

	gs.ui.PrettyPrint(r.Flatted())
	return nil
}

// ListAll lists all collection's records
func (gs *Gostore) ListAll() error {
	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	r, err := gs.store.ReadAll()
	if err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Flatted()...)
	return nil
}

// Search the collection for records matching given query. Query follows
// blevesearch syntax (https://blevesearch.com/docs/Query-String-Query/).
func (gs *Gostore) Search(query string) error {
	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	r, err := gs.store.Search(query)
	if err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Flatted()...)
	return nil
}

// Edit updates an existing record from the collection
func (gs *Gostore) Edit(key string) error {
	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	r, err := gs.store.Read(key)
	if err != nil {
		return err
	}

	mdata, err := gs.ui.Edit(r.UserValue())
	if err != nil {
		return err
	}
	r.ReplaceValue(mdata)

	if err := modules.ProcessRecord(r, gs.updateModules); err != nil {
		return err
	}

	if !gs.pretend {
		if err := gs.store.Update(key, r); err != nil {
			return err
		}
	}

	gs.ui.PrettyPrint(r.Flatted())
	return nil
}

// Delete removes a record from the collection.
func (gs *Gostore) Delete(key string) error {
	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	if !gs.pretend {
		if err := gs.store.Delete(key); err != nil {
			return err
		}
	}

	return nil
}

// Export copies a record's media file from the collection to the given destination.
func (gs *Gostore) Export(key, dstFolder string) (err error) {
	dstPath := filepath.Join(key, dstFolder)

	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	f, err := gs.store.OpenRecord(key)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return err
	}

	var w *os.File
	w, err = os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer func() {
		cerr := w.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(w, f)
	return
}

// CheckAndRepair verifies collection's consistency and repairs or reports found inconsistencies.
func (gs *Gostore) CheckAndRepair() error {
	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	orphans, err := gs.store.CheckAndRepair()
	if err != nil {
		return err
	}

	if len(orphans) > 0 {
		//TODO: use PrettyPrint to show list of orphans?
		gs.ui.Printf("Found orphans files in the collection:\n%s\n", strings.Join(orphans, "\n"))
	}
	return nil
}
