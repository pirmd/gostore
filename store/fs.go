package store

import (
	"os"
	"path/filepath"

	"github.com/pirmd/gostore/store/vfs"
	"github.com/pirmd/gostore/util"
)

type storefs struct {
	path        string
	validNameFn func(string) bool
	fs          *vfs.VFS
}

func newFS(path string, validFn func(string) bool) *storefs {
	return &storefs{
		path:        path,
		validNameFn: validFn,
	}
}

// Open opens a new fs for use.
func (s *storefs) Open() error {
	if err := os.MkdirAll(s.path, 0777); err != nil {
		return err
	}

	s.fs = vfs.NewJailfs(s.path, vfs.NewFilterfs(s.validNameFn, vfs.NewOsfs()))
	return nil
}

// Close cleanly closes a storefs.
func (s *storefs) Close() error {
	return nil
}

// Exists checks whether a path is present in the storefs.
func (s *storefs) Exists(path string) (bool, error) {
	return s.fs.Exists(path)
}

// Put imports a Record into the storefs.
// Put will happily erase and replace any existing file previously
// found at Record's path, if any.
func (s *storefs) Put(r *Record) error {
	return s.fs.Copy(r.File(), r.Key())
}

// Get returns a vfs.File for reading to the Record's file corresponding to the
// given key.
// If Record's Key is absolute, store will look for Record's content from the
// host file-system, other wise it get it from store's storage.
func (s *storefs) Get(key string) (ReadCloser, error) {
	return s.fs.Open(key)
}

// Update updates an existing Record in the storefs.
func (s *storefs) Update(oldkey string, r *Record) error {
	if r.File() != nil {
		s.Put(r)
		if oldkey != r.Key() {
			s.Delete(oldkey)
		}
	}

	return s.Move(oldkey, r)
}

// Move moves a Record in the storefs.
func (s *storefs) Move(oldkey string, r *Record) error {
	return s.fs.Move(oldkey, r.Key())
}

// Delete removes a record from the storefs as well as its parent directories if empty
// If key does not exist, Delete exits without error.
func (s *storefs) Delete(key string) error {
	if err := s.fs.Remove(key); err != nil {
		return err
	}

	for {
		key = filepath.Dir(key)
		if err := s.fs.Remove(key); err != nil {
			break
		}
	}
	return nil
}

// Walk iterates over all storefs items and call walkFn for each item. Errors
// that happen during walkFn execution will not stop the execution of Walk but
// are captured and will be returned once Walk is over.
func (s *storefs) Walk(walkFn func(string) error) error {
	errWalk := new(util.MultiErrors)

	if errw := s.fs.Walk("", func(path string, info os.FileInfo, err error) error {
		if err == os.ErrPermission {
			return filepath.SkipDir
		}
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if err := walkFn(filepath.ToSlash(path)); err != nil {
			errWalk.Add(err)
		}
		return nil
	}); errw != nil {
		return errw
	}

	return errWalk.Err()
}

// SearchGlob looks for the records from storefs whose path matches the given pattern.
// Under the hood the pattern matching follows the same behaviour than filepath.Match.
func (s *storefs) SearchGlob(pattern string) (matches []string, err error) {
	err = s.Walk(func(path string) error {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return err
		}
		if matched {
			matches = append(matches, path)
		}
		return nil
	})
	return
}
