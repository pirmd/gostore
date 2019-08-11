package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/processing"
	"github.com/pirmd/gostore/store"
)

type gostore struct {
	root string
	ui   UserInterfacer
}

func newGostore() *gostore {
	configure() //XXX: clean all of that once config has ben restructured

	return &gostore{
		root: cfg.StoreRoot,
		ui:   NewUI(cfg.UIAuto),
	}
}

func (gs *gostore) open() (*store.Store, error) {
	//TODO: introduce a read-ony / dry-run / pretend mode)
	return store.Open(
		gs.root,
		store.UsingLogger(debugger),
		store.UsingTypeField(media.TypeField),
	)
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

	mdata, err := gs.ui.Merge(mdataFromFile, mdataFetched)
	if err != nil {
		return err
	}

	//XXX: change path to filepath.Base(path)?
	r := store.NewRecord(path, mdata)

	if err := processing.ProcessRecord(r); err != nil {
		return err
	}

	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	if err := collection.Create(r, f); err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Fields())
	return nil
}

//Info retrieves information about any collection's record.
//If fromFile flag is set, Info also displays actual metadata stored in the
//media file
func (gs *gostore) Info(key string, fromFile bool) error {
	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	r, err := collection.Read(key)
	if err != nil {
		return err
	}

	if fromFile {
		f, err := collection.OpenRecord(key)
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
func (gs *gostore) ListAll() error {
	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	r, err := collection.ReadAll()
	if err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Fields()...)
	return nil
}

// Search the collection for records matching given query. Query follows
// blevesearch syntax (https://blevesearch.com/docs/Query-String-Query/).
func (gs *gostore) Search(query string) error {
	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	r, err := collection.Search(query)
	if err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Fields()...)
	return nil
}

// Edit updates an existing record from the collection
func (gs *gostore) Edit(key string) error {
	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	r, err := collection.Read(key)
	if err != nil {
		return err
	}

	mdata, err := gs.ui.Edit(r.OrigValue())
	if err != nil {
		return err
	}
	r.ReplaceValues(mdata)

	if err := processing.ProcessRecord(r); err != nil {
		return err
	}

	if err := collection.Update(key, r); err != nil {
		return err
	}

	gs.ui.PrettyPrint(r.Fields())
	return nil
}

// Delete removes a record from the collection.
func (gs *gostore) Delete(key string) error {
	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	if err := collection.Delete(key); err != nil {
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

	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	r, err := collection.OpenRecord(key)
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
	collection, err := gs.open()
	if err != nil {
		return err
	}
	defer collection.Close()

	orphans, err := collection.CheckAndRepair()
	if err != nil {
		return err
	}

	if len(orphans) > 0 {
		//TODO: use PrettyPrint to show list of orphans?
		gs.ui.Printf("Found orphans files in the collection:\n%s\n", strings.Join(orphans, "\n"))
	}
	return nil
}
