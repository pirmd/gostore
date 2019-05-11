package store

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

const (
	BucketName = "gostore"
)

var (
	ErrRecordNotFoundInDb = fmt.Errorf("Record not found")
)

type storedb struct {
	path string
	db   *bolt.DB
}

//newDB creates a new database
func newDB(path string) *storedb {
	return &storedb{path: path}
}

//Open opens a new database.
func (s *storedb) Open() (err error) {
	if s.db, err = bolt.Open(s.path, 0666, nil); err != nil {
		return
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(BucketName))
		return err
	})
}

//Close closes database
func (s *storedb) Close() error {
	return s.db.Close()
}

//Put adds a new Record into the database
func (s *storedb) Put(r *Record) error {
	buf, err := json.Marshal(r.value)
	if err != nil {
		return err
	}

	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		return b.Put([]byte(r.key), buf)
	}); err != nil {
		return err
	}

	return nil
}

//Delete remove a Record entry from the database
func (s *storedb) Delete(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		return b.Delete([]byte(key))
	})
}

//Get retrieves a value from the database and unmarshal it
func (s *storedb) Get(key string) (*Record, error) {
	var buf []byte
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		buf = b.Get([]byte(key))
		return nil
	}); err != nil {
		return nil, err
	}

	if buf == nil {
		return nil, ErrRecordNotFoundInDb
	}

	r := NewRecord(key, nil)
	if err := json.Unmarshal(buf, &r.value); err != nil {
		return nil, err
	}
	return r, nil
}

//Exists check if a value is in the storedb
func (s *storedb) Exists(key string) (bool, error) {
	var buf []byte

	//TODO: check if something better than Get can be used
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		buf = b.Get([]byte(key))
		return nil
	}); err != nil {
		return false, err
	}

	return (buf != nil), nil
}

//Walk iterates over all storedb items and call walkFn for each item
//Walk does not stop if an error is reported by walkFn, such errors will
//be captured and reported back once Walk is over
func (s *storedb) Walk(walkFn func(string) error) error {
	errWalk := new(NonBlockingErrors)

	//boltdb does not support modifying the database during
	//iteration, so we first get the all keys, then we act on
	//them

	keys := []string{}
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		return b.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	}); err != nil {
		return err
	}

	for _, k := range keys {
		if err := walkFn(k); err != nil {
			errWalk.Add(err)
		}
	}

	return errWalk.Err()
}
