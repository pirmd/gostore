package vfs

import (
	"io"
	"os"
	"path/filepath"
)

// Exists checks if a file or directory exists
func (vfs *VFS) Exists(path string) (bool, error) {
	_, err := vfs.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// Copy copies a Reader content to a file within the vfs.
// dst file and its parent folders are create if they don't exist yet
// Copy does not prevent nor warn if dst is already existing
func (vfs *VFS) Copy(src io.Reader, dst string) (err error) {
	err = vfs.MkdirAll(filepath.Dir(dst), 0777)
	if err != nil {
		return
	}

	var w File
	w, err = vfs.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	defer func() {
		cerr := w.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(w, src)
	return
}

// Move moves a file or folder to a new location. If destination
// is in a non existing path, Move creates it first.
func (vfs *VFS) Move(src, dst string) error {
	if err := vfs.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}

	return vfs.Rename(src, dst)
}
