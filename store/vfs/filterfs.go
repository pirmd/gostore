package vfs

import (
	"os"
	"time"
)

//The filterfs filters access to files according to a validating function (func(string) bool).
//Any file with a path, for which the validating function returns false, will not be allowed.
//Typical use is with a regexp validating function (using regexp.MatchString() for example)
//
//filterfs is not a standalone filesystem as it passes all operation to an underlying filesystem,
//it only checks that the given path is not filtered before cascading it to its wrapped file-system
//
//Symlinks are not recognized
type filterfs struct {
	wrappedfs *VFS
	validfn   func(string) bool
}

//Newfilterfs creates a new filterfs that relies on a validating function to
//allow or not access to file-system files depending on their name.
//
//The validating function should, if appropriate, make sure that the forbidden
//name can be embedded in a full path name (filterfs does not take care of that).
//For example, disallowing access to hidden files should consider ".name", "folder/.name"
//and "folder/.name/file"
func NewFilterfs(validfn func(string) bool, wrappedfs *VFS) *VFS {
	return &VFS{
		&filterfs{
			validfn:   validfn,
			wrappedfs: wrappedfs,
		},
	}
}

func (fs *filterfs) Mkdir(name string, mode os.FileMode) error {
	if fs.validfn(name) {
		return fs.wrappedfs.Mkdir(name, mode)
	}
	return os.ErrPermission
}

func (fs *filterfs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	if fs.validfn(name) {
		f, err := fs.wrappedfs.OpenFile(name, flag, perm)
		if err != nil {
			return nil, err
		}
		return &filterFile{f, fs.validfn}, nil
	}
	return nil, os.ErrPermission
}

func (fs *filterfs) Remove(name string) error {
	if fs.validfn(name) {
		return fs.wrappedfs.Remove(name)
	}
	return os.ErrPermission
}

func (fs *filterfs) Rename(oldname, newname string) error {
	if fs.validfn(oldname) && fs.validfn(newname) {
		return fs.wrappedfs.Rename(oldname, newname)
	}
	return os.ErrPermission
}

func (fs *filterfs) Stat(name string) (os.FileInfo, error) {
	if fs.validfn(name) {
		return fs.wrappedfs.Stat(name)
	}
	return nil, os.ErrPermission
}

//Lstat is here to fulfill Fs interface but we don't allow to follow symlink here
func (fs *filterfs) Lstat(name string) (os.FileInfo, error) {
	return fs.Stat(name)
}

func (fs *filterfs) Chmod(name string, mode os.FileMode) error {
	if fs.validfn(name) {
		return fs.wrappedfs.Chmod(name, mode)
	}
	return os.ErrPermission
}

func (fs *filterfs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if fs.validfn(name) {
		return fs.wrappedfs.Chtimes(name, atime, mtime)
	}
	return os.ErrPermission
}

type filterFile struct {
	File
	validfn func(string) bool
}

func (f *filterFile) Readdir(count int) ([]os.FileInfo, error) {
	childFi, err := f.File.Readdir(count)
	if err != nil {
		return nil, err
	}

	var res []os.FileInfo
	for _, fi := range childFi {
		if f.validfn(fi.Name()) {
			res = append(res, fi)
		}
	}
	return res, nil
}
