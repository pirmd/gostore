package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
	"github.com/pirmd/gostore/ui"
	"github.com/pirmd/gostore/ui/cli"
	"github.com/pirmd/gostore/util"
)

// Gostore represents the main collection manager.
type Gostore struct {
	log           *log.Logger
	debug         *log.Logger
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
		log:   log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		debug: log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),

		pretend:       cfg.ReadOnly,
		deleteGhosts:  cfg.DeleteGhosts,
		deleteOrphans: cfg.DeleteOrphans,
		importOrphans: cfg.ImportOrphans,
	}

	if cfg.Verbose || cfg.Debug {
		gs.log.SetOutput(os.Stderr)
	}

	if cfg.Debug {
		gs.debug.SetOutput(os.Stderr)
		cfg.Store.Logger = gs.debug
	}

	var err error
	if gs.store, err = store.NewFromConfig(cfg.Store); err != nil {
		return nil, err
	}

	if gs.ui, err = cli.NewFromConfig(cfg.UI); err != nil {
		return nil, err
	}

	env := &modules.Environment{Logger: gs.log, UI: gs.ui, Store: gs.store}
	for _, module := range cfg.Import {
		m, err := modules.New(module.Name, module.Config, env)
		if err != nil {
			return nil, fmt.Errorf("cannot create module '%s': %v", module.Name, err)
		}

		gs.importModules = append(gs.importModules, m)
	}

	for _, module := range cfg.Update {
		m, err := modules.New(module.Name, module.Config, env)
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

// ListAll lists all collection's records.
func (gs *Gostore) ListAll(sortBy []string) error {
	r, err := gs.store.ReadAll()
	if err != nil {
		return err
	}

	gs.ui.PrettyPrint(util.Sort(r.Flatted(), sortBy)...)
	return nil
}

// ListGlob retrieves information about a collection's record.
func (gs *Gostore) ListGlob(pattern []string, sortBy []string) error {
	r, err := gs.glob(pattern)
	if err != nil {
		return fmt.Errorf("listing '%s' failed: %s", pattern, err)
	}

	if len(r) != 0 {
		gs.ui.PrettyPrint(util.Sort(r.Flatted(), sortBy)...)
	}

	return nil
}

// ListQuery searches the collection for records matching given query. Query
// follows bleve's search syntax (https://blevesearch.com/docs/Query-String-Query/).
func (gs *Gostore) ListQuery(query string, sortBy []string) error {
	r, err := gs.store.ReadQuery(query)
	if err != nil {
		return err
	}

	gs.ui.PrettyPrint(util.Sort(r.Flatted(), sortBy)...)
	return nil
}

// Edit updates an existing record from the collection
func (gs *Gostore) Edit(pattern []string) error {
	records, err := gs.glob(pattern)
	if err != nil {
		return fmt.Errorf("editing '%s' failed: %s", pattern, err)
	}

	var editErr util.MultiErrors
	for _, r := range records {
		gs.log.Printf("Editing '%s'", r.Key())

		mdata, err := gs.ui.Edit(r.Data())
		if err != nil {
			editErr.Add(fmt.Errorf("editing '%s' failed: %s", r.Key(), err))
			continue
		}

		if err := gs.update(r, mdata); err != nil {
			editErr.Add(fmt.Errorf("editing '%s' failed: %s", r.Key(), err))
			continue
		}

		records = append(records, r)
	}

	if len(records) != 0 {
		gs.ui.PrettyPrint(records.Flatted()...)
	}

	return editErr.Err()
}

// MultiEdit updates a set of records at once.
func (gs *Gostore) MultiEdit(pattern []string) error {
	records, err := gs.glob(pattern)
	if err != nil {
		return fmt.Errorf("editing '%s' failed: %s", pattern, err)
	}

	if len(records) == 0 {
		return nil
	}

	mdata, err := gs.ui.MultiEdit(records.Data())
	if err != nil {
		return fmt.Errorf("editing '%s' failed: %s", pattern, err)
	}

	if len(mdata) != len(records) {
		return fmt.Errorf("editing '%s' failed: number of records after multi-edition is inconsistent", pattern)
	}

	var editErr util.MultiErrors
	for i := range mdata {
		if err := gs.update(records[i], mdata[i]); err != nil {
			editErr.Add(fmt.Errorf("editing '%s' failed: %s", records[i].Key(), err))
			continue
		}
	}

	gs.ui.PrettyPrint(records.Flatted()...)

	return editErr.Err()
}

// Delete removes a record from the collection.
func (gs *Gostore) Delete(pattern []string) error {
	records, err := gs.glob(pattern)
	if err != nil {
		return fmt.Errorf("deleting '%s' failed: %s", pattern, err)
	}

	var delErr util.MultiErrors
	for _, r := range records {
		gs.log.Printf("Deleting '%s'", r.Key())

		if !gs.pretend {
			if err := gs.store.Delete(r.Key()); err != nil {
				delErr.Add(fmt.Errorf("deleting '%s' failed: %s", r.Key(), err))
				continue
			}
		}
	}

	return delErr.Err()
}

// Export copies a record's media file from the collection to the given destination.
func (gs *Gostore) Export(dstFolder string, pattern []string) error {
	records, err := gs.glob(pattern)
	if err != nil {
		return fmt.Errorf("exporting '%s' failed: %s", pattern, err)
	}

	var exportErr util.MultiErrors
	for _, r := range records {
		gs.log.Printf("Exporting '%s'", r.Key())

		if err := gs.export(r, dstFolder); err != nil {
			exportErr.Add(fmt.Errorf("exporting '%s' failed: %s", r.Key(), err))
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
			if err := gs.Delete(ghosts); err != nil {
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
			if err := gs.Delete(orphans); err != nil {
				errCheck.Add(err)
			}

		case gs.importOrphans:
			if err := gs.Edit(orphans); err != nil {
				errCheck.Add(err)
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

// RebuildIndex deletes then rebuild the index from scratch based on the
// database content. It can be used for example to implement a new mapping
// strategy or if things are really going bad
func (gs *Gostore) RebuildIndex() error {
	if err := gs.store.RebuildIndex(); err != nil {
		return fmt.Errorf("rebuilding index failed: %s", err)
	}
	return nil
}

// Fields lists fields names that are available for search or for templates.
// Some fields might only be available for a given media Type.
func (gs *Gostore) Fields() error {
	fields, err := gs.store.Fields()
	if err != nil {
		return fmt.Errorf("listing collection's known fields failed: %s", err)
	}
	gs.ui.Printf("%s\n", strings.Join(fields, "\n"))
	return nil
}

func (gs *Gostore) glob(pattern []string) (store.Records, error) {
	var rec store.Records

	for _, p := range pattern {
		r, err := gs.store.ReadGlob(p)
		if err != nil {
			return nil, fmt.Errorf("looking for '%s' failed: %s", p, err)
		}

		rec = append(rec, r...)
	}

	return rec, nil
}

func (gs *Gostore) insert(path string) (*store.Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := store.NewRecord(filepath.Base(path), nil)
	r.SetFile(f)

	if err := modules.ProcessRecord(r, gs.importModules); err != nil {
		return nil, err
	}

	if !gs.pretend {
		if err := gs.store.Insert(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (gs *Gostore) update(r *store.Record, mdata map[string]interface{}) error {
	key := r.Key()

	r.SetData(mdata)

	if err := modules.ProcessRecord(r, gs.updateModules); err != nil {
		return err
	}

	if !gs.pretend {
		if err := gs.store.Update(key, r); err != nil {
			return err
		}
	}

	return nil
}

func (gs *Gostore) export(r *store.Record, dstFolder string) (err error) {
	dstPath := filepath.Join(dstFolder, r.Key())

	f, err := gs.store.OpenRecord(r)
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
