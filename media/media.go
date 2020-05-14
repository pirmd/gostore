package media

import (
	"errors"
	"io"
	"os"

	"github.com/pirmd/gostore/store"
)

var (
	// ErrNoMetadataFound reports an error when no Metadata found
	ErrNoMetadataFound = errors.New("media: no metadata found")
)

// Metadata represents a set of media's metadata, it is essentially a set of (key,
// value).
type Metadata = store.Value

// File represents a media file
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// GetMetadata reads metadata from the provided File and setup the proper media
// Type if not done by the corresponding Handler.
func GetMetadata(f File) (Metadata, error) {
	mh, err := handlers.ForReader(f)
	if err != nil {
		return nil, err
	}

	mdata, err := mh.GetMetadata(f)
	if err != nil {
		return nil, err
	}

	mdata.Set(TypeField, mh.Type())

	return mdata, nil
}

// GetMetadataFromFile reads metadata from the provided filename
func GetMetadataFromFile(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return GetMetadata(f)
}
