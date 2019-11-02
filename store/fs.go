package store

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pirmd/gostore/store/vfs"
)

type storefs struct {
	path        string
	validNameFn func(string) bool
	fs          *vfs.VFS
}

//newFS returns a new storefs
func newFS(path string, validFn func(string) bool) *storefs {
	return &storefs{
		path:        path,
		validNameFn: validFn,
	}
}

//Open opens a new fs for use
func (s *storefs) Open() error {
	if err := os.MkdirAll(s.path, 0777); err != nil {
		return err
	}

	s.fs = vfs.NewJailfs(s.path, vfs.NewFilterfs(s.validNameFn, vfs.NewOsfs()))
	return nil
}

//Close cleanly closes a storefs
func (s *storefs) Close() error {
	return nil
}

//Exists checks whether a path is present in the storefs
func (s *storefs) Exists(path string) (bool, error) {
	return s.fs.Exists(path)
}

//Put imports a Record into the storefs
//
//Put will happily erase and replace any existing file previously
//found at Record's path, if any.
func (s *storefs) Put(r *Record, src io.Reader) error {
	if err := s.fs.Import(src, r.key); err != nil {
		return err
	}
	return nil
}

//Get returns an os.File fore reading/wroting to the Record's file
//corresponding to the gicen key
func (s *storefs) Get(path string) (vfs.File, error) {
	return s.fs.Open(path)
}

//Move moves a Record in the storefs
func (s *storefs) Move(oldpath string, r *Record) error {
	if err := s.fs.Move(oldpath, r.key); err != nil {
		return err
	}
	return nil
}

//Delete removes a record from the storefs as well as its parent directories if empty
//If path does not exist, Delete exits without error
func (s *storefs) Delete(path string) error {
	if err := s.fs.Remove(path); err != nil {
		return err
	}

	for {
		path = filepath.Dir(path)
		if err := s.fs.Remove(path); err != nil {
			break
		}
	}
	return nil
}

//Walk iterates over all storedb items and call f for each item
//Errors that happen during f execution will not stop the execution of Walk but
//are captured and will be returned once Walk is over
func (s *storefs) Walk(walkFn func(string) error) error {
	errWalk := new(NonBlockingErrors)

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
