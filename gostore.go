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
	"github.com/pirmd/gostore/util"
)

// Gostore represents the main collection manager.
type Gostore struct {
	log           *log.Logger
	pretend       bool
	deleteGhosts  bool
	deleteOrphans bool
	importOrphans bool
	store         *store.Store
	ui            ui.UserInterfacer
	importModules []modules.Module
	updateModules []modules.Module
}

func newGostore(cfg *Config) (*Gostore, error) {
	cfg.expandEnv()

	gs := &Gostore{
		log:           log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		pretend:       cfg.ReadOnly,
		deleteGhosts:  cfg.DeleteGhosts,
		deleteOrphans: cfg.DeleteOrphans,
		importOrphans: cfg.ImportOrphans,
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

	for _, module := range cfg.Import {
		m, err := modules.New(module.Name, module.Config, gs.log, gs.ui)
		if err != nil {
			return nil, fmt.Errorf("cannot create module '%s': %v", module.Name, err)
		}

		gs.importModules = append(gs.importModules, m)
	}

	for _, module := range cfg.Update {
		m, err := modules.New(module.Name, module.Config, gs.log, gs.ui)
		if err != nil {
			return nil, fmt.Errorf("cannot create module '%s': %v", module.Name, err)
		}

		gs.updateModules = append(gs.updateModules, m)
	}

	return gs, nil
}

func openGostore(cfg *Config) (*Gostore, error) {
	gs, err := newGostore(cfg)
	if err != nil {
		return nil, err
	}

	if err := gs.Open(); err != nil {
		return nil, err
	}

	return gs, nil
}

// Open opens a gostore for read/write operation
func (gs *Gostore) Open() error {
	if err := gs.store.Open(); err != nil {
		return fmt.Errorf("opening gostore failed: %s", err)
	}

	return nil
}

// Close cleanly closes a gostore
func (gs *Gostore) Close() error {
	return gs.store.Close()
}

// Import inserts new media into the collection
func (gs *Gostore) Import(mediaFiles []string) error {
	var newRecords store.Records
	var importErr util.MultiErrors

	for _, path := range mediaFiles {
		gs.log.Printf("Importing '%s'", path)

		r, err := gs.insert(path)
		if err != nil {
			importErr.Add(fmt.Errorf("importing '%s' failed: %s", path, err))
			continue
		}

		newRecords = append(newRecords, r)
	}

	if len(newRecords) != 0 {
		gs.ui.PrettyPrint(newRecords.Flatted()...)
	}

	return importErr.Err()
}

// List retrieves information about a collection's record.
func (gs *Gostore) List(pattern ...string) error {
	var records store.Records
	var listErr util.MultiErrors

	for _, p := range pattern {
		r, err := gs.store.ReadGlob(p)
		if err != nil {
			listErr.Add(fmt.Errorf("listing '%s' failed: %s", p, err))
			continue
		}

		if len(r) == 0 {
			listErr.Add(fmt.Errorf("listing '%s' failed: no such record", p))
			continue
		}

		records = append(records, r...)
	}

	if len(records) != 0 {
		gs.ui.PrettyPrint(records.Flatted()...)
	}

	return listErr.Err()
}

// ListAll lists all collection's records
func (gs *Gostore) ListAll() error {
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
	r, err := gs.store.Search(query)
	if err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Flatted()...)
	return nil
}

// Edit updates an existing record from the collection
func (gs *Gostore) Edit(key string) error {
	r, err := gs.store.Read(key)
	if err != nil {
		return err
	}

	mdata, err := gs.ui.Edit(r.Data())
	if err != nil {
		return err
	}
	r.SetData(mdata)

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
func (gs *Gostore) Delete(pattern ...string) error {
	var keys []string
	var delErr util.MultiErrors

	for _, p := range pattern {
		k, err := gs.store.SearchKeys(p)
		if err != nil {
			delErr.Add(fmt.Errorf("deleting '%s' failed: %s", p, err))
			continue
		}

		if len(k) == 0 {
			delErr.Add(fmt.Errorf("deleting '%s' failed: no such record", p))
			continue
		}

		keys = append(keys, k...)
	}

	for _, key := range keys {
		gs.log.Printf("Deleting '%s'", key)

		if !gs.pretend {
			if err := gs.store.Delete(key); err != nil {
				delErr.Add(fmt.Errorf("deleting '%s' failed: %s", key, err))
				continue
			}
		}
	}

	return delErr.Err()
}

// Export copies a record's media file from the collection to the given destination.
func (gs *Gostore) Export(dstFolder string, pattern ...string) error {
	var keys []string
	var exportErr util.MultiErrors

	for _, p := range pattern {
		k, err := gs.store.SearchKeys(p)
		if err != nil {
			exportErr.Add(fmt.Errorf("exporting '%s' failed: %s", p, err))
			continue
		}

		if len(k) == 0 {
			exportErr.Add(fmt.Errorf("exporting '%s' failed: no such record", p))
			continue
		}

		keys = append(keys, k...)
	}

	for _, key := range keys {
		gs.log.Printf("Exporting '%s'", key)

		if err := gs.export(key, dstFolder); err != nil {
			exportErr.Add(fmt.Errorf("exporting '%s' failed: %s", key, err))
			continue
		}
	}

	return exportErr.Err()
}

// CheckAndRepair verifies collection's consistency and repairs or reports
// found inconsistencies.
// Behaviour in case of inconsistencies depends on gostore's DeleteGhosts,
// DeleteOrphans or ImportOrphans flags.
func (gs *Gostore) CheckAndRepair() error {
	var errCheck util.MultiErrors

	ghosts, err := gs.store.CheckGhosts()
	if err != nil {
		errCheck.Add(err)
	}
	if len(ghosts) > 0 {
		switch {
		case gs.deleteGhosts:
			if err := gs.Delete(ghosts...); err != nil {
				errCheck.Add(err)
			}

		default:
			gs.ui.Printf("Found ghosts records in the collection:\n%s\n", strings.Join(ghosts, "\n"))
		}
	}

	orphans, err := gs.store.CheckOrphans()
	if err != nil {
		errCheck.Add(err)
	}
	if len(orphans) > 0 {
		switch {
		case gs.deleteOrphans:
			if err := gs.Delete(orphans...); err != nil {
				errCheck.Add(err)
			}

		case gs.importOrphans:
			for _, o := range orphans {
				if err := gs.Edit(o); err != nil {
					errCheck.Add(err)
				}
			}

		default:
			gs.ui.Printf("Found orphans files in the collection:\n%s\n", strings.Join(orphans, "\n"))
		}
	}

	if err := gs.store.RepairIndex(); err != nil {
		errCheck.Add(err)
	}

	return errCheck.Err()
}

func (gs *Gostore) insert(path string) (*store.Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mdataFromFile, err := media.ReadMetadata(f)
	if err != nil {
		return nil, err
	}
	r := store.NewRecord(filepath.Base(path), mdataFromFile)

	if err := modules.ProcessRecord(r, gs.importModules); err != nil {
		return nil, err
	}

	if !gs.pretend {
		if err := gs.store.Create(r, f); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (gs *Gostore) export(key, dstFolder string) (err error) {
	dstPath := filepath.Join(dstFolder, key)

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
