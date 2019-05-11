package vfs

import (
	"io"
	"os"
	"time"
)

//VFS encapsulates operations provided to interact with the file-system
//
//Nota: VFS functions are to be found in filepath.go and fsutil.go when they are
//"ported" from the standard library (path/filepath and os respectively).
//util.go contains extra functions that can be helpful
type VFS struct {
	filesystem
}

//File represents a file
//
//File provides a stripped down interface not as complete as os.File but
//are enough for my use cases. For that reason, some vfs provided in this
//package are not supercharged to provide correct feedback for os.File func
//that are not in File (for example Readdirnames is not adapted for filterfs
//or jailfs)
type File interface {
	io.Closer
	io.Seeker
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt

	//Readdir reads the contents of the directory associated with file and
	//returns a slice of up to n FileInfo values, as would be returned by
	//os.File.Readdir
	Readdir(count int) ([]os.FileInfo, error)
}

//fileSystem represents a File-System
type filesystem interface {
	//Mkdir creates a directory in the files-ystem
	Mkdir(name string, perm os.FileMode) error

	//OpenFile opens a file using the given flags and the given mode.
	OpenFile(name string, flag int, perm os.FileMode) (File, error)

	//Remove removes a file identified by name
	Remove(name string) error

	//Rename renames a file.
	Rename(oldname, newname string) error

	//Stat returns a FileInfo describing the file
	Stat(name string) (os.FileInfo, error)

	//Lstat returns a FileInfo describing the file. It follows symlink if any and if supported
	Lstat(name string) (os.FileInfo, error)

	//Chmod changes the mode of the file.
	Chmod(name string, mode os.FileMode) error

	//Chtimes changes the access and modification times of the file
	Chtimes(name string, atime time.Time, mtime time.Time) error
}
