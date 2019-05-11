package vfs

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	tstCases = []string{
		"file.txt",
		"folder",
		"folder/file.txt",
		"folder/subfolder",
		"folder/subfolder/file.txt",
	}
)

func PopulateFs(fs *VFS, root string, tree []string) error {
	for _, f := range tree {
		path := filepath.Join(root, f)
		if filepath.Ext(path) == "" {
			if err := fs.MkdirAll(path, 0777); err != nil {
				return fmt.Errorf("Fail to create temporary folder '%s': %v", path, err)
			}
		} else {
			if _, err := fs.Create(path); err != nil {
				return fmt.Errorf("Fail to create temporary file '%s': %v", path, err)
			}
		}
	}
	return nil
}

//returns list of relative path inside root
func ListFs(fs *VFS, root string) (files []string, err error) {
	err = fs.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		relpath, _ := filepath.Rel(root, path)
		files = append(files, relpath)
		return nil
	})
	return
}
