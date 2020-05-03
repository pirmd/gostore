package main

import (
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

	// ImportModules lists of processings to be applied when importing a record
	ImportModules map[string]*rawYAMLConfig

	// UpdateModules lists of processings to be applied when updating a record
	UpdateModules map[string]*rawYAMLConfig
}

func newConfig() *Config {
	return &Config{
		//XXX: needed?
		Store: &store.Config{Path: "."},
		UI:    &cli.Config{},
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
	return cfg.unmarshal(v)
}

// Gostore represents the main collection manager.
type Gostore struct {
	log *log.Logger
	//XXX: rename pretend to readonly
	pretend       bool
	store         *store.Store
	ui            ui.UserInterfacer
	importModules []modules.Module
	updateModules []modules.Module
}

//XXX: like ui / Store: have gostoreNew and gostoreNewFromConfig
//XXX: like ui / Store: have Modules: New and NewFromConfig
func newGostore(cfg *Config) (*Gostore, error) {
	gs := &Gostore{
		log:     log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		pretend: cfg.ReadOnly,
	}

	if cfg.ShowLog {
		gs.log.SetOutput(os.Stderr)
	}

	gs.pretend = cfg.ReadOnly

	var err error
	if gs.store, err = store.NewFromConfig(cfg.Store); err != nil {
		return nil, err
	}

	if gs.ui, err = cli.NewFromConfig(cfg.UI); err != nil {
		return nil, err
	}

	for modName, modRawCfg := range cfg.ImportModules {
		m, err := modules.New(modName, modRawCfg, gs.log)
		if err != nil {
			return nil, err
		}

		gs.importModules = append(gs.importModules, m)
	}

	for modName, modRawCfg := range cfg.UpdateModules {
		m, err := modules.New(modName, modRawCfg, gs.log)
		if err != nil {
			return nil, err
		}

		gs.updateModules = append(gs.updateModules, m)
	}

	return gs, nil
}

// Import adds a new media into the collection
func (gs *Gostore) Import(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	mdataFromFile, err := media.GetMetadata(f)
	if err != nil {
		return err
	}

	mdataFetched, err := media.FetchMetadata(mdataFromFile)
	if err != nil && err != media.ErrNoMetadataFound {
		return err
	}

	mdata, err := gs.ui.Merge(mdataFromFile, mdataFetched)
	if err != nil {
		return err
	}

	r := store.NewRecord(filepath.Base(path), mdata)

	if err := modules.ProcessRecord(r, gs.importModules); err != nil {
		return err
	}

	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	if !gs.pretend {
		if err := gs.store.Create(r, f); err != nil {
			return err
		}
	}

	gs.ui.PrettyPrint(r.Fields())
	return nil
}

// Info retrieves information about any collection's record.
// If fromFile flag is set, Info also displays actual metadata stored in the
// media file
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

		gs.ui.PrettyDiff(r.OrigValue(), mdata)
		return nil
	}

	gs.ui.PrettyPrint(r.Fields())
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

	gs.ui.PrettyPrint(r.Fields()...)
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

	gs.ui.PrettyPrint(r.Fields()...)
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

	mdata, err := gs.ui.Edit(r.OrigValue())
	if err != nil {
		return err
	}
	r.ReplaceValues(mdata)

	if err := modules.ProcessRecord(r, gs.updateModules); err != nil {
		return err
	}

	if !gs.pretend {
		if err := gs.store.Update(key, r); err != nil {
			return err
		}
	}

	gs.ui.PrettyPrint(r.Fields())
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
// Destination is considered as a folder where the record's media file will be
// copied to (final file's will be dst/key, keeping any sub-folder(s) coming
// with the record's key name).
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
