package vfs

import (
	"os"
	"path/filepath"
	"sort"
)

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func (vfs *VFS) Walk(root string, walkFn filepath.WalkFunc) error {
	info, err := vfs.Lstat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = walk(vfs, root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

//copied from https://golang.org/src/path/filepath/path.go
func walk(vfs *VFS, path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readDirNames(vfs, path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := vfs.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(vfs, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

//copied from https://golang.org/src/path/filepath/path.go
//with small adpatation to accomodate Readdir instead of Readdirnames
func readDirNames(vfs *VFS, dirname string) ([]string, error) {
	f, err := vfs.Open(dirname)
	if err != nil {
		return nil, err
	}

	fi, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, i := range fi {
		names = append(names, i.Name())
	}

	sort.Strings(names)
	return names, nil
}
