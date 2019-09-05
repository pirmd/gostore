package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

// Gostore represents the main collection manager.
type Gostore struct {
	log           *log.Logger
	store         *store.Store
	ui            UserInterfacer
	importModules []modules.Module
	updateModules []modules.Module
}

//XXX: default IMportModule
func newGostore(cfg *config) (*Gostore, error) {
	gs := &Gostore{
		log: log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile),
		ui:  NewCLI(cfg.UI),
	}

	if cfg.ShowLog {
		gs.log.SetOutput(os.Stderr)
	}

	var err error
	gs.store, err = store.New(
		//XXX: check what is happening if Root is empty (expected behaviour: use current dir)
		cfg.Store.Root,
		store.UsingLogger(gs.log),
		store.UsingTypeField(media.TypeField),
	)
	if err != nil {
		return nil, err
	}

	for _, modName := range cfg.ImportModules {
		m, err := modules.New(modName, cfg.Modules[modName], gs.log)
		if err != nil {
			return nil, err
		}

		gs.importModules = append(gs.importModules, m)
	}

	for _, modName := range cfg.UpdateModules {
		m, err := modules.New(modName, cfg.Modules[modName], gs.log)
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

	//XXX: change path to filepath.Base(path)?
	r := store.NewRecord(path, mdata)

	if err := modules.ProcessRecord(r, gs.importModules); err != nil {
		return err
	}

	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	if err := gs.store.Create(r, f); err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Fields())
	return nil
}

//Info retrieves information about any collection's record.
//If fromFile flag is set, Info also displays actual metadata stored in the
//media file
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

//ListAll lists all collection's records
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

	if err := gs.store.Update(key, r); err != nil {
		return err
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

	if err := gs.store.Delete(key); err != nil {
		return err
	}

	return nil
}

// Export copies a record's media file from the collection to the given destination.
// Destination is considered as a folder where the record's media file will be
// copied to (final file's will be dst/key, keeping any sub-folder(s) coming
// with the record's key name).
func (gs *Gostore) Export(key, dstFolder string) error {
	dstPath := filepath.Join(key, dstFolder)

	if err := gs.store.Open(); err != nil {
		return err
	}
	defer gs.store.Close()

	r, err := gs.store.OpenRecord(key)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return err
	}

	w, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer w.Close() //XXX: check for failures on close

	if _, err := io.Copy(w, r); err != nil {
		//XXX: is it going to work? search correct pattern
		_ = os.Remove(dstPath)
		return err
	}

	return nil
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

func init() {
	//XXX: or in config.ModuleConifg(name string) -> Module ??
	modules.ConfUnmarshal = yaml.Unmarshal
}
