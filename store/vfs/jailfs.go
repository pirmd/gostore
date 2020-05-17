package vfs

import (
	"os"
	"path/filepath"
	"time"
)

// jailfs restricts all operations to a given path within a filesystem.
//
// Name given to the jailfs operation are considered as relative to the jail's
// root (example: ../test1 -> root/test1) Therefore, any file name pointing
// outside the jail root path will most of the time found to be non existing
// file.
//
// Operation that modifies the root (delete, move, chmod, chtime) are not
// allowed from the jailfs.
//
// jailfs is not a standalone vfs.filesystem as it passes all operation to an
// underlying vfs.filesystem, it only checks that the cascaded path is within
// the Jail and/or considers input path as relative to the Jail's root.
//
// jailfs is designed as a convenient way to work inside a folder it does not
// pretend to be a secured way to jail any application.
type jailfs struct {
	root      string
	wrappedfs *VFS
}

// NewJailfs creates a new composite Fs It does not check that root exists nor
// create it if it doesn't
func NewJailfs(root string, wrappedfs *VFS) *VFS {
	return &VFS{
		&jailfs{
			root:      filepath.Clean(root),
			wrappedfs: wrappedfs,
		},
	}
}

func (fs *jailfs) Mkdir(name string, mode os.FileMode) error {
	path := fs.realPath(name)
	if path == fs.root {
		return os.ErrPermission
	}
	return fs.wrappedfs.Mkdir(path, mode)
}

func (fs *jailfs) OpenFile(name string, flag int, mode os.FileMode) (File, error) {
	return fs.wrappedfs.OpenFile(fs.realPath(name), flag, mode)
}

func (fs *jailfs) Remove(name string) error {
	path := fs.realPath(name)
	if path == fs.root {
		return os.ErrPermission
	}
	return fs.wrappedfs.Remove(path)
}

func (fs *jailfs) Rename(oldname, newname string) error {
	oldpath := fs.realPath(oldname)
	if oldpath == fs.root {
		return os.ErrPermission
	}
	newpath := fs.realPath(newname)
	if newpath == fs.root {
		return os.ErrPermission
	}
	return fs.wrappedfs.Rename(oldpath, newpath)
}

func (fs *jailfs) Stat(name string) (os.FileInfo, error) {
	return fs.wrappedfs.Stat(fs.realPath(name))
}

// Lstat is here to fulfill Fs interface but we don't allow to follow symlink
// here
func (fs *jailfs) Lstat(name string) (os.FileInfo, error) {
	return fs.Stat(name)
}

func (fs *jailfs) Chmod(name string, mode os.FileMode) error {
	path := fs.realPath(name)
	if path == fs.root {
		return os.ErrPermission
	}
	return fs.wrappedfs.Chmod(path, mode)
}

func (fs *jailfs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	path := fs.realPath(name)
	if path == fs.root {
		return os.ErrPermission
	}
	return fs.wrappedfs.Chtimes(path, atime, mtime)
}

// realPath returns the "real" path of a file within a jail. Path are "secured"
// to some point by ignoring any indication pointing outside of the Jail's root.
//
// realPath does not check if the 'real' path exists or not or makes sense, it
// is the responsibility of client func to detect anything surprising.
func (fs *jailfs) realPath(path string) string {
	return filepath.Join(fs.root, filepath.Clean("/"+path))
}
