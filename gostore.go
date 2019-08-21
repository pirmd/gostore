package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

type gostore struct {
	//XXX: switch all to private
	Store         *store.Store
	UI            UserInterfacer
	ImportModules []modules.Module
	UpdateModules []modules.Module
}

func newGostore(cfg *gostoreConfig) (*gostore, error) {
	logger := cfg.Log.Logger()

	s, err := store.New(
		cfg.Store.Root,
		store.UsingLogger(logger),
		store.UsingTypeField(media.TypeField),
	)
	if err != nil {
		return nil, err
	}

	var impMods []modules.Module
	for _, modName := range cfg.ImportModules {
		//XXX: test empty config (nil)
		m, err := modules.New(modName, cfg.Modules[modName], logger)
		if err != nil {
			return nil, err
		}

		impMods = append(impMods, m)
	}

	var upMods []modules.Module
	for _, modName := range cfg.UpdateModules {
		//XXX: test empty config (nil)
		m, err := modules.New(modName, cfg.Modules[modName], logger)
		if err != nil {
			return nil, err
		}

		upMods = append(upMods, m)
	}

	return &gostore{
		Store: s,
		UI:    NewUI(cfg.UI),

		ImportModules: impMods,
		UpdateModules: upMods,
	}, nil
}

// Import adds a new media into the collection
func (gs *gostore) Import(path string) error {
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

	mdata, err := gs.UI.Merge(mdataFromFile, mdataFetched)
	if err != nil {
		return err
	}

	//XXX: change path to filepath.Base(path)?
	r := store.NewRecord(path, mdata)

	if err := modules.ProcessRecord(r, gs.ImportModules); err != nil {
		return err
	}

	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	if err := gs.Store.Create(r, f); err != nil {
		return err
	}

	gs.UI.PrettyPrint(r.Fields())
	return nil
}

//Info retrieves information about any collection's record.
//If fromFile flag is set, Info also displays actual metadata stored in the
//media file
func (gs *gostore) Info(key string, fromFile bool) error {
	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	r, err := gs.Store.Read(key)
	if err != nil {
		return err
	}

	if fromFile {
		f, err := gs.Store.OpenRecord(key)
		if err != nil {
			return err
		}
		defer f.Close()

		mdata, err := media.GetMetadata(f)
		if err != nil {
			return err
		}

		gs.UI.PrettyDiff(r.OrigValue(), mdata)
		return nil
	}

	gs.UI.PrettyPrint(r.Fields())
	return nil
}

//ListAll lists all collection's records
func (gs *gostore) ListAll() error {
	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	r, err := gs.Store.ReadAll()
	if err != nil {
		return err
	}

	gs.UI.PrettyPrint(r.Fields()...)
	return nil
}

// Search the collection for records matching given query. Query follows
// blevesearch syntax (https://blevesearch.com/docs/Query-String-Query/).
func (gs *gostore) Search(query string) error {
	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	r, err := gs.Store.Search(query)
	if err != nil {
		return err
	}

	gs.UI.PrettyPrint(r.Fields()...)
	return nil
}

// Edit updates an existing record from the collection
func (gs *gostore) Edit(key string) error {
	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	r, err := gs.Store.Read(key)
	if err != nil {
		return err
	}

	mdata, err := gs.UI.Edit(r.OrigValue())
	if err != nil {
		return err
	}
	r.ReplaceValues(mdata)

	if err := modules.ProcessRecord(r, gs.UpdateModules); err != nil {
		return err
	}

	if err := gs.Store.Update(key, r); err != nil {
		return err
	}

	gs.UI.PrettyPrint(r.Fields())
	return nil
}

// Delete removes a record from the collection.
func (gs *gostore) Delete(key string) error {
	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	if err := gs.Store.Delete(key); err != nil {
		return err
	}

	return nil
}

// Export copies a record's media file from the collection to the given destination.
// Destination is considered as a folder where the record's media file will be
// copied to (final file's will be dst/key, keeping any sub-folder(s) coming
// with the record's key name).
func (gs *gostore) Export(key, dstFolder string) error {
	dstPath := filepath.Join(key, dstFolder)

	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	r, err := gs.Store.OpenRecord(key)
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

// Repair verifies collection's consistency and repairs or reports found inconsistencies.
func (gs *gostore) CheckAndRepair() error {
	if err := gs.Store.Open(); err != nil {
		return err
	}
	defer gs.Store.Close()

	orphans, err := gs.Store.CheckAndRepair()
	if err != nil {
		return err
	}

	if len(orphans) > 0 {
		//TODO: use PrettyPrint to show list of orphans?
		gs.UI.Printf("Found orphans files in the collection:\n%s\n", strings.Join(orphans, "\n"))
	}
	return nil
}
