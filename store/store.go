package store

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pirmd/gostore/store/vfs"
)

const (
	dbPath  = ".store_database.db"
	idxPath = ".store_index"
)

var (
	ErrKeyIsNotValid       = fmt.Errorf("key is invalid.")
	ErrRecordAlreadyExists = fmt.Errorf("record already exists.")
	ErrRecordDoesNotExist  = fmt.Errorf("record does not exist.")
)

//Store represents the actual storing engine
//It is made of a filesystem, a keystore (leveldb) and an indexer (bleve)
type Store struct {
	fs  *storefs
	db  *storedb
	idx *storeidx
}

//New creates a new Store
//New accepts options to costumize default Store behaviour
func New(path string, opts ...option) (*Store, error) {
	s := &Store{}

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

//Open creates and opens a Store for use
//It is a simple shortcut for s := New() then s.Open(). It accepts the same Options than New()
func Open(path string, opts ...option) (s *Store, err error) {
	logger.Printf("Opening store at %s", path)
	if s, err = New(path, opts...); err != nil {
		return
	}

	err = s.Open()
	return
}

//Open opens a Store for use
func (s *Store) Open() error {
	logger.Printf("Opening store's filesystem")
	if err := s.fs.Open(); err != nil {
		return err
	}

	logger.Printf("Opening store's database")
	if err := s.db.Open(); err != nil {
		if e := s.fs.Close(); e != nil {
			err = fmt.Errorf("%s\nClose store's filesystem failed: %s", err, e)
		}
		return err
	}

	logger.Printf("Opening store's Index")
	if err := s.idx.Open(); err != nil {
		if e := s.fs.Close(); e != nil {
			err = fmt.Errorf("%s\nClose store's filesystem failed: %s", err, e)
		}
		if e := s.db.Close(); e != nil {
			err = fmt.Errorf("%s\nClose store's database failed: %s", err, e)
		}
		s.db.Close()
		return err
	}

	return nil
}

//Close cleanly closes a Store
func (s *Store) Close() error {
	err := new(NonBlockingErrors)

	logger.Printf("Closing store's filesystem")
	if e := s.fs.Close(); e != nil {
		err.Add(fmt.Errorf("Fail to close store's filesystem: %s", e))
	}

	logger.Printf("Closing store's database")
	if e := s.db.Close(); e != nil {
		err.Add(fmt.Errorf("Fail to close store's database: %s", e))
	}

	logger.Printf("Closing store's index")
	if e := s.idx.Close(); e != nil {
		err.Add(fmt.Errorf("Fail to close store's index: %s", e))
	}

	return err.Err()
}

//Create creates a new record in the Store
//Create does not replace existing Record (you have to use Update for that) but will
//replace partially existing records resulting from an inconsistent state of the Store
//(e.g. file exists but entry in db does not)
func (s *Store) Create(r *Record, file io.Reader) error {
	logger.Printf("Adding new record to store '%s'", r)

	exists, err := s.Exists(r.key)
	if err != nil {
		return err
	}
	if exists {
		return ErrRecordAlreadyExists
	}

	logger.Printf("Register new record in store's db")
	if err := s.db.Put(r); err != nil {
		return err
	}

	logger.Printf("Register new record in store's idx")
	if err := s.idx.Put(r); err != nil {
		if e := s.db.Delete(r.key); e != nil {
			err = fmt.Errorf("%s\nFail to clean db after error: %s", err, e)
		}

		return err
	}

	logger.Printf("Import new media file into store's fs")
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

//Exists returns whether a Record exists for the given key
//If the Store's state is inconsistent for the given key (e.g. file is not present but
//and entry exists in the database), Exists returns false
func (s *Store) Exists(key string) (exists bool, err error) {
	logger.Printf("Test if '%s' exists in fs", key)
	if exists, err = s.fs.Exists(key); err != nil || !exists {
		return
	}

	logger.Printf("Test if '%s' exists in db", key)
	if exists, err = s.db.Exists(key); err != nil || !exists {
		return
	}

	logger.Printf("Test if '%s' exists in idx", key)
	return s.idx.Exists(key)
}

//Read returns the stored Record corresponding to the given key
func (s *Store) Read(key string) (*Record, error) {
	logger.Printf("Get record '%s' from storage", key)
	return s.db.Get(key)
}

//ReadAll returns all store's records
func (s *Store) ReadAll() (list Records, err error) {
	logger.Printf("Get all records from store")

	err = s.db.Walk(func(key string) error {
		r, err := s.db.Get(key)
		if err != nil {
			return err
		}
		list = append(list, r)
		return nil
	})

	return
}

//OpenRecord opens an os.File for the record corresponding to the given key
func (s *Store) OpenRecord(key string) (vfs.File, error) {
	logger.Printf("Open record '%s' from storage", key)
	return s.fs.Get(key)
}

//Update replaces an existing Store's record.
//Update only works if the record is actually existing in the Store. If it
//corresponds to a partially existing Record (e.g. no file but data in the
//database), Update will fail (you have to use Create for that situation)
func (s *Store) Update(key string, r *Record) error {
	logger.Printf("Updating record '%s' to '%s'", key, r)

	exists, err := s.Exists(key)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRecordDoesNotExist
	}

	if r.key != key {
		exists, err := s.Exists(r.key)
		if err != nil {
			return err
		}
		if exists {
			return ErrRecordAlreadyExists
		}
	}

	logger.Printf("Updating record in store's db")
	if err := s.db.Put(r); err != nil {
		return err
	}

	logger.Printf("Updating record in store's idx")
	if err := s.idx.Put(r); err != nil {
		if e := s.db.Delete(r.key); e != nil {
			err = fmt.Errorf("%s\nFail to clean db after error: %s", err, e)
		}
		return err
	}

	if r.key != key {
		logger.Printf("Import new media file into store's fs")
		if err := s.fs.Move(key, r); err != nil {
			if e := s.db.Delete(r.key); e != nil {
				err = fmt.Errorf("%s\nFail to clean db after error: %s", err, e)
			}

			if e := s.idx.Delete(r.key); e != nil {
				err = fmt.Errorf("%s\nFail to clean idx after error: %s", err, e)
			}

			return err
		}

		errDel := new(NonBlockingErrors)

		logger.Printf("Clean old entry '%s' in the store's db", key)
		if err := s.db.Delete(key); err != nil {
			errDel.Add(fmt.Errorf("Fail to clean db from old entry: %s", err))
		}

		logger.Printf("Clean old entry '%s' in the store's idx", key)
		if err := s.idx.Delete(key); err != nil {
			errDel.Add(fmt.Errorf("Fail to clean idx from old entry: %s", err))
		}

		return errDel.Err()
	}

	return nil
}

//Delete removes a record from the Store
func (s *Store) Delete(key string) error {
	logger.Printf("Deleting record '%s' from store", key)

	logger.Printf("Deleting record's file from store's fs")
	if err := s.fs.Delete(key); err != nil {
		return err
	}

	errDel := new(NonBlockingErrors)

	logger.Printf("Deleting record from store's db")
	if err := s.db.Delete(key); err != nil {
		errDel.Add(fmt.Errorf("Fail to clean db from old entry: %s", err))
	}

	logger.Printf("Deleting record from store's idx")
	if err := s.idx.Delete(key); err != nil {
		errDel.Add(fmt.Errorf("Fail to clean idx from old entry: %s", err))
	}

	return errDel.Err()
}

//Search returns the list of keys corresponding to the given search query
//The query should follow the bleve search engine synthax.
func (s *Store) Search(query string) (Records, error) {
	logger.Printf("Search records for '%s'", query)

	keys, err := s.idx.Search(query)
	if err != nil {
		return nil, err
	}

	var result Records
	for _, key := range keys {
		r, err := s.db.Get(key)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

//Rebuild deletes then rebuild the index from scratch based on the database content
//It can be used for example to implement a new mapping stratégy or if things are
//realy going bad
func (s *Store) RebuildIndex() error {
	logger.Printf("Create a new index from scratch")
	if err := s.idx.Empty(); err != nil {
		return err
	}

	logger.Printf("Rebuilding index")
	return s.db.Walk(func(key string) error {
		r, err := s.db.Get(key)
		if err != nil {
			return err
		}
		return s.idx.Put(r)
	})
}

//CheckAndRepair verifies the consistency between the Store's components (file system, database
//and index) and try to solve any issues:
// - delete database and index entries whose file cannot be found
// - re-create index entry with a database and file record
// - report files found in the store without any index or database record
func (s *Store) CheckAndRepair() ([]string, error) {
	errCheck := new(NonBlockingErrors)

	logger.Printf("Verify that all store's database entries are in the store")
	if err := s.db.Walk(func(key string) error {
		logger.Printf("Verify record '%s' from database...", key)
		exists, err := s.fs.Exists(key)
		if err != nil {
			return err
		}
		if !exists {
			logger.Printf("Record '%s' is in database and not in filesystem. Deleting it", key)
			return s.Delete(key)
		}

		exists, err = s.idx.Exists(key)
		if err != nil {
			return err
		}
		if !exists {
			logger.Printf("Record '%s' is in database and not in index. Adding it to index", key)
			r, err := s.db.Get(key)
			if err != nil {
				return err
			}
			return s.idx.Put(r)
		}

		return nil
	}); err != nil {
		errCheck.Add(err)
	}
	logger.Printf("Verify that all store's files are in the store database")
	orphans := []string{}
	if err := s.fs.Walk(func(key string) error {
		logger.Printf("Verify record '%s' from filesystem...", key)
		exists, err := s.db.Exists(key)
		if err != nil {
			return err
		}
		if !exists {
			logger.Printf("File '%s' is not in database. Either delete it or add it to store", key)
			orphans = append(orphans, key)
		}
		return nil
	}); err != nil {
		errCheck.Add(err)
	}

	logger.Printf("Verify that all indexed records are in the store's database")
	if err := s.idx.Walk(func(key string) error {
		logger.Printf("Verify record '%s' from index...", key)
		exists, err := s.db.Exists(key)
		if err != nil {
			return err
		}
		if !exists {
			logger.Printf("Record '%s' is indexed and is not in the store's database. Deleting it from index", key)
			return s.idx.Delete(key)
		}
		return nil
	}); err != nil {
		errCheck.Add(err)
	}

	return orphans, errCheck.Err()
}

func (s *Store) isValidKey(key string) bool {
	cleanKey := filepath.ToSlash(filepath.Clean("/" + key))[1:]

	return cleanKey != "/" &&
		!strings.HasPrefix(cleanKey, dbPath) &&
		!strings.HasPrefix(cleanKey, idxPath)
}
