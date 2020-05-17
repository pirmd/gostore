package vfs

import (
	"os"
	"time"
)

// osfs is a vfs.filesystem implementation that give access to the underlying os
// file-system It basically wraps-up "os" package to meet vfs.filesystem
// interface
type osfs struct{}

// NewOsfs creates a vfs for the underlying os file-system
func NewOsfs() *VFS {
	return &VFS{&osfs{}}
}

func (fs *osfs) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (fs *osfs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

func (fs *osfs) Remove(name string) error {
	return os.Remove(name)
}

func (fs *osfs) Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}

func (fs *osfs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *osfs) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func (fs *osfs) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

func (fs *osfs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}
