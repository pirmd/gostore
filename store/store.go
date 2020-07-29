package store

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/pirmd/gostore/store/vfs"
	"github.com/pirmd/gostore/util"
)

const (
	dbPath  = ".store_database.db"
	idxPath = ".store_index"
)

var (
	// ErrKeyIsNotValid raises an error if provided key is invalid
	ErrKeyIsNotValid = errors.New("key is invalid")

	// ErrRecordAlreadyExists raises an error if a record already exits
	ErrRecordAlreadyExists = errors.New("record already exists")

	// ErrRecordDoesNotExist raises an error is a record does not exist
	ErrRecordDoesNotExist = errors.New("record does not exist")
)

// Store represents the actual storing engine. It is made of a filesystem, a
// key-value database and an indexer (bleve)
type Store struct {
	fs  *storefs
	db  *storedb
	idx *storeidx

	log *log.Logger
}

// New creates a new Store. New accepts options to customize default Store
// behaviour
func New(path string, opts ...Option) (*Store, error) {
	s := &Store{
		log: log.New(ioutil.Discard, "store ", log.LstdFlags),
	}

	s.fs = newFS(path, s.isValidKey)
	s.db = newDB(filepath.Join(path, dbPath))
	s.idx = newIdx(filepath.Join(path, idxPath))

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Open opens a Store for use
func (s *Store) Open() error {
	s.log.Printf("Opening store's filesystem")
	if err := s.fs.Open(); err != nil {
		return err
	}

	s.log.Printf("Opening store's database")
	if err := s.db.Open(); err != nil {
		if e := s.fs.Close(); e != nil {
			err = fmt.Errorf("%s\nClose store's filesystem failed: %s", err, e)
		}
		return err
	}

	s.log.Printf("Opening store's Index")
	if err := s.idx.Open(); err != nil {
		if e := s.fs.Close(); e != nil {
			err = fmt.Errorf("%s\nClose store's filesystem failed: %s", err, e)
		}
		if e := s.db.Close(); e != nil {
			err = fmt.Errorf("%s\nClose store's database failed: %s", err, e)
		}
		return err
	}

	return nil
}

// Close cleanly closes a Store
func (s *Store) Close() error {
	err := new(util.MultiErrors)

	s.log.Printf("Closing store's filesystem")
	if e := s.fs.Close(); e != nil {
		err.Add(fmt.Errorf("fail to close store's filesystem: %s", e))
	}

	s.log.Printf("Closing store's database")
	if e := s.db.Close(); e != nil {
		err.Add(fmt.Errorf("fail to close store's database: %s", e))
	}

	s.log.Printf("Closing store's index")
	if e := s.idx.Close(); e != nil {
		err.Add(fmt.Errorf("fail to close store's index: %s", e))
	}

	return err.Err()
}

// Create creates a new record in the Store. Create does not replace existing
// Record (you have to use Update for that) but will replace partially existing
// records resulting from an inconsistent state of the Store (e.g. file exists
// but entry in db does not)
func (s *Store) Create(r *Record, file io.Reader) error {
	s.log.Printf("Adding new record to store '%s'", r.Key())

	exists, err := s.Exists(r.key)
	if err != nil {
		return err
	}
	if exists {
		return ErrRecordAlreadyExists
	}

	s.log.Printf("Register new record in store's db")
	if err := s.db.Put(r); err != nil {
		return err
	}

	s.log.Printf("Register new record in store's idx")
	if err := s.idx.Put(r); err != nil {
		if e := s.db.Delete(r.key); e != nil {
			err = fmt.Errorf("%s\nFail to clean db after error: %s", err, e)
		}

		return err
	}

	s.log.Printf("Import new media file into store's fs")
	if err := s.fs.Put(r, file); err != nil {
		if e := s.db.Delete(r.key); e != nil {
			err = fmt.Errorf("%s\nFail to clean db after error: %s", err, e)
		}

		if e := s.idx.Delete(r.key); e != nil {
			err = fmt.Errorf("%s\nFail to clean idx after error: %s", err, e)
		}

		return err
	}

	return nil
}

// Exists returns whether a Record exists for the given key. If the Store's state
// is inconsistent for the given key (e.g. file is not present but and entry
// exists in the database), Exists returns false
func (s *Store) Exists(key string) (exists bool, err error) {
	s.log.Printf("Test if '%s' exists in fs", key)
	if exists, err = s.fs.Exists(key); err != nil || !exists {
		return
	}

	s.log.Printf("Test if '%s' exists in db", key)
	if exists, err = s.db.Exists(key); err != nil || !exists {
		return
	}

	s.log.Printf("Test if '%s' exists in idx", key)
	return s.idx.Exists(key)
}

// Read returns the stored Record corresponding to the given key
func (s *Store) Read(key string) (*Record, error) {
	s.log.Printf("Get record '%s' from storage", key)
	r, err := s.db.Get(key)
	if err != nil {
		return nil, err
	}

	// Retrieving data from store's database relies on JSON unmarshaling to a
	// map[string]interface{} type. It will miss detecting properly time and
	// numeric values. At this point using information stored in the store's
	// index can help at this kind of information is found there.
	ridx, err := s.idx.Get(key)
	if err != nil {
		return nil, err
	}

	for k, v := range ridx.Data() {
		switch v := v.(type) {
		case time.Time, float64:
			r.SetIfExists(k, v)
		}
	}

	return r, nil
}

// ReadAll returns all store's records
func (s *Store) ReadAll() (list Records, err error) {
	s.log.Printf("Get all records from store")

	err = s.db.Walk(func(key string) error {
		r, err := s.Read(key)
		if err != nil {
			return err
		}
		list = append(list, r)
		return nil
	})

	return
}

// ReadGlob returns the list of records corresponding to the given search pattern.
// Under the hood the pattern matching follows the same behaviour than
// filepath.Match.
func (s *Store) ReadGlob(pattern string) (Records, error) {
	s.log.Printf("Get all records that match glob '%s' from store", pattern)

	keys, err := s.SearchGlob(pattern)
	if err != nil {
		return nil, err
	}

	var result Records
	for _, key := range keys {
		r, err := s.Read(key)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

// ReadQuery returns the Records that match the given search query. The query
// follows the bleve search engine syntax (http://blevesearch.com/docs/Query-String-Query/).
func (s *Store) ReadQuery(query string) (Records, error) {
	s.log.Printf("Rearch records that match query '%s'", query)

	keys, err := s.SearchQuery(query)
	if err != nil {
		return nil, err
	}

	var result Records
	for _, key := range keys {
		r, err := s.Read(key)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

// OpenRecord opens the record corresponding to the given key for reading.
func (s *Store) OpenRecord(key string) (vfs.File, error) {
	s.log.Printf("Open record '%s' from storage", key)
	return s.fs.Get(key)
}

// SearchQuery returns the Records' keys that match the given search query. The
// query and sort order should follow the bleve search engine syntax.
func (s *Store) SearchQuery(query string) ([]string, error) {
	s.log.Printf("Search records for '%s'", query)

	keys, err := s.idx.Search(query)
	if err != nil {
		return nil, err
	}

	s.log.Printf("Found records: %+v", keys)
	return keys, nil
}

// SearchGlob returns the Records' keys that match the given glob pattern.
// Under the hood the pattern matching follows the same behaviour than
// filepath.Match.
func (s *Store) SearchGlob(pattern string) ([]string, error) {
	s.log.Printf("Search records for ID='%s'", pattern)

	keys, err := s.fs.Search(pattern)
	if err != nil {
		return nil, err
	}

	s.log.Printf("Found records: %+v", keys)
	return keys, nil
}

// SearchFields returns the Records' keys that match the provided fields value.
// Level of accepted fuzziness can be specified.
func (s *Store) SearchFields(fields map[string]interface{}, fuzziness int) ([]string, error) {
	s.log.Printf("Search records for FIELDS='%#v' with FUZZY=%d", fields, fuzziness)

	keys, err := s.idx.SearchFields(fields, fuzziness)
	if err != nil {
		return nil, err
	}

	s.log.Printf("Found records: %+v", keys)
	return keys, nil
}

// Update replaces an existing Store's record. Update will fail if the new
// record key is already existing (ErrRecordAlreadyExists).
func (s *Store) Update(key string, r *Record) error {
	s.log.Printf("Updating record '%s' to '%s'", key, r.Key())

	if r.key != key {
		exists, err := s.Exists(r.key)
		if err != nil {
			return err
		}
		if exists {
			return ErrRecordAlreadyExists
		}
	}

	s.log.Printf("Updating record in store's db")
	if err := s.db.Put(r); err != nil {
		return err
	}

	s.log.Printf("Updating record in store's idx")
	if err := s.idx.Put(r); err != nil {
		if e := s.db.Delete(r.key); e != nil {
			err = fmt.Errorf("%s\nFail to clean db after error: %s", err, e)
		}
		return err
	}

	if r.key != key {
		s.log.Printf("Import new media file into store's fs")
		if err := s.fs.Move(key, r); err != nil {
			if e := s.db.Delete(r.key); e != nil {
				err = fmt.Errorf("%s\nFail to clean db after error: %s", err, e)
			}

			if e := s.idx.Delete(r.key); e != nil {
				err = fmt.Errorf("%s\nFail to clean idx after error: %s", err, e)
			}

			return err
		}

		errDel := new(util.MultiErrors)

		s.log.Printf("Clean old entry '%s' in the store's db", key)
		if err := s.db.Delete(key); err != nil {
			errDel.Add(fmt.Errorf("fail to clean db from old entry: %s", err))
		}

		s.log.Printf("Clean old entry '%s' in the store's idx", key)
		if err := s.idx.Delete(key); err != nil {
			errDel.Add(fmt.Errorf("fail to clean idx from old entry: %s", err))
		}

		return errDel.Err()
	}

	return nil
}

// Delete removes a record from the Store
func (s *Store) Delete(key string) error {
	s.log.Printf("Deleting record '%s' from store", key)

	errDel := new(util.MultiErrors)

	s.log.Printf("Deleting record's file from store's fs")
	if err := s.fs.Delete(key); err != nil {
		errDel.Add(fmt.Errorf("fail to remove old entry: %s", err))
	}

	s.log.Printf("Deleting record from store's db")
	if err := s.db.Delete(key); err != nil {
		errDel.Add(fmt.Errorf("fail to clean db from old entry: %s", err))
	}

	s.log.Printf("Deleting record from store's idx")
	if err := s.idx.Delete(key); err != nil {
		errDel.Add(fmt.Errorf("fail to clean idx from old entry: %s", err))
	}

	return errDel.Err()
}

// RebuildIndex deletes then rebuild the index from scratch based on the
// database content. It can be used for example to implement a new mapping
// strategy or if things are really going bad
func (s *Store) RebuildIndex() error {
	s.log.Printf("Create a new index from scratch")
	if err := s.idx.Empty(); err != nil {
		return err
	}

	errRebuild := new(util.MultiErrors)
	s.log.Printf("Rebuilding index")
	return s.db.Walk(func(key string) error {
		r, err := s.db.Get(key)
		if err != nil {
			errRebuild.Add(err)
			return nil
		}
		if err := s.idx.Put(r); err != nil {
			errRebuild.Add(err)
		}

		return errRebuild.Err()
	})
}

// RepairIndex check the consistency between the index and the database. Try to
// repair them as far as possible.
func (s *Store) RepairIndex() error {
	var errRepair util.MultiErrors

	s.log.Printf("Verify that all store's database entries are in the store's index")
	if err := s.db.Walk(func(key string) error {
		exists, err := s.idx.Exists(key)
		if err != nil {
			errRepair.Add(err)
			return nil
		}
		if !exists {
			s.log.Printf("Record '%s' is in database and not in index. Adding it to index", key)
			r, err := s.db.Get(key)
			if err != nil {
				errRepair.Add(err)
				return nil
			}
			if err := s.idx.Put(r); err != nil {
				errRepair.Add(err)
			}
			return nil
		}

		return nil
	}); err != nil {
		errRepair.Add(err)
	}

	s.log.Printf("Verify that all indexed records are in the store's database")
	if err := s.idx.Walk(func(key string) error {
		exists, err := s.db.Exists(key)
		if err != nil {
			return err
		}
		if !exists {
			s.log.Printf("Record '%s' is indexed and is not in the store's database. Deleting it from index", key)
			return s.idx.Delete(key)
		}
		return nil
	}); err != nil {
		errRepair.Add(err)
	}

	return errRepair.Err()
}

// CheckGhosts lists any database entries that has no corresponding file
// in the store's filesystem.
func (s *Store) CheckGhosts() ([]string, error) {
	var orphans []string

	s.log.Printf("Verify that all store's database entries are in the store")
	err := s.db.Walk(func(key string) error {
		exists, err := s.fs.Exists(key)
		if err != nil {
			return err
		}
		if !exists {
			s.log.Printf("Record '%s' is in database but not in filesystem", key)
			orphans = append(orphans, key)
		}
		return nil
	})

	return orphans, err
}

// CheckOrphans Lists any file in store that is not in the database.
func (s *Store) CheckOrphans() ([]string, error) {
	var orphans []string

	s.log.Printf("Verify that all store's files are in the store database")
	err := s.fs.Walk(func(key string) error {
		exists, err := s.db.Exists(key)
		if err != nil {
			return err
		}
		if !exists {
			s.log.Printf("File '%s' is not in database", key)
			orphans = append(orphans, key)
		}
		return nil
	})

	return orphans, err
}

// IsDirty verify store's general health in term of consistency between
// database, index and filesystem.
func (s *Store) IsDirty() bool {
	s.log.Printf("Verify that all store's database records are in the filesystem and index")
	if err := s.db.Walk(func(key string) error {
		exists, err := s.fs.Exists(key)
		if err != nil || !exists {
			return err
		}
		exists, err = s.idx.Exists(key)
		if err != nil || !exists {
			return err
		}
		return nil
	}); err != nil {
		return true
	}

	s.log.Printf("Verify that all store's files are in the store database")
	if err := s.fs.Walk(func(key string) error {
		exists, err := s.db.Exists(key)
		if err != nil || !exists {
			return err
		}
		return nil
	}); err != nil {
		return true
	}

	s.log.Printf("Verify that all indexed records are in the store's database")
	if err := s.idx.Walk(func(key string) error {
		exists, err := s.db.Exists(key)
		if err != nil || !exists {
			return err
		}
		return nil
	}); err != nil {
		return true
	}

	return false
}

// Fields list the fields that can be used when searching the collection.
func (s *Store) Fields() ([]string, error) {
	return s.idx.Fields()
}

func (s *Store) isValidKey(key string) bool {
	cleanKey := filepath.ToSlash(filepath.Clean("/" + key))[1:]

	return cleanKey != "/" &&
		!strings.HasPrefix(cleanKey, dbPath) &&
		!strings.HasPrefix(cleanKey, idxPath)
}
