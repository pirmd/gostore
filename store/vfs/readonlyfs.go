package vfs

import (
	"os"
	"time"
)

// The readonlyfs restricts all operations that modify in way or another the
// filesystem.
//
// readonlyfs is not a standalone vfs.filesystem as it passes all read
// operation to an underlying vfs.filesystem. Should a write operation being
// asked, an os.ErrPermission is raised.
type readonlyfs struct {
	wrappedfs *VFS
}

// NewReadonlyfs creates a new composite read-only Fs.
func NewReadonlyfs(wrappedfs *VFS) *VFS {
	return &VFS{
		&readonlyfs{
			wrappedfs: wrappedfs,
		},
	}
}

func (fs *readonlyfs) Mkdir(name string, mode os.FileMode) error {
	return os.ErrPermission
}

func (fs *readonlyfs) OpenFile(name string, flag int, mode os.FileMode) (File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, os.ErrPermission
	}

	return fs.wrappedfs.OpenFile(name, flag, mode)
}

func (fs *readonlyfs) Remove(name string) error {
	return os.ErrPermission
}

func (fs *readonlyfs) Rename(oldname, newname string) error {
	return os.ErrPermission
}

func (fs *readonlyfs) Stat(name string) (os.FileInfo, error) {
	return fs.wrappedfs.Stat(name)
}

func (fs *readonlyfs) Lstat(name string) (os.FileInfo, error) {
	return fs.wrappedfs.Lstat(name)
}

func (fs *readonlyfs) Chmod(name string, mode os.FileMode) error {
	return os.ErrPermission
}

func (fs *readonlyfs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.ErrPermission
}
