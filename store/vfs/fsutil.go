package vfs

import (
	"os"
	"path/filepath"
	"syscall"
)

// Create creates the named file mode 0666 (before umask) on the given
// filesystem, truncating it if it already exists.  The associated file
// descriptor has mode os.O_RDWR.  If there is an error, it will be of type
// *os.PathError.
//
// "ported" from standard lib os.Create
func (vfs *VFS) Create(name string) (File, error) {
	return vfs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Open opens the named file on the given filesystem for reading.
// If successful, methods on the returned file can be used for reading.
// The associated file descriptor has mode os.O_RDONLY.
// If there is an error, it will be of type *PathError.
//
// "ported" from standard lib os.Open
func (vfs *VFS) Open(name string) (File, error) {
	return vfs.OpenFile(name, os.O_RDONLY, 0)
}

// MkdirAll creates a directory named path on the given filesystem,
// along with any necessary parents, and returns nil,
// or else returns an error.
// The permission bits perm are used for all
// directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing
// and returns nil.
//
//"ported" from standard lib os.MkdirAll
func (vfs *VFS) MkdirAll(path string, perm os.FileMode) error {
	dir, err := vfs.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}

	i := len(path)
	for i > 0 && os.IsPathSeparator(path[i-1]) {
		i--
	}

	j := i
	for j > 0 && !os.IsPathSeparator(path[j-1]) {
		j--
	}

	if j > 1 {
		err = vfs.MkdirAll(path[0:j-1], perm)
		if err != nil {
			return err
		}
	}

	err = vfs.Mkdir(path, perm)
	if err != nil {
		dir, err1 := vfs.Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return err
	}
	return nil
}

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters.  If the path does not exist, RemoveAll
// returns nil.
//
// "ported" from standard lib os.RemoveAll
func (vfs *VFS) RemoveAll(path string) error {
	err := vfs.Remove(path)
	if err == nil || os.IsNotExist(err) {
		return nil
	}

	dir, serr := vfs.Lstat(path)
	if serr != nil {
		if serr, ok := serr.(*os.PathError); ok && (os.IsNotExist(serr.Err) || serr.Err == syscall.ENOTDIR) {
			return nil
		}
		return serr
	}
	if !dir.IsDir() {

		return err
	}

	fd, err := vfs.Open(path)
	if err != nil {
		return err
	}
	list, err := fd.Readdir(-1)
	if err != nil {
		return err
	}

	err = nil
	for _, fi := range list {
		err1 := vfs.RemoveAll(filepath.Join(path, fi.Name()))
		if err == nil {
			err = err1
		}
	}

	err1 := vfs.Remove(path)
	if err1 == nil || os.IsNotExist(err1) {
		return nil
	}
	if err == nil {
		err = err1
	}
	return err
}
