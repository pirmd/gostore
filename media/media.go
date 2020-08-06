package media

import (
	"errors"
	"io"
	"os"
)

var (
	// ErrNoMetadataFound reports an error when no Metadata found
	ErrNoMetadataFound = errors.New("media: no metadata found")
)

// Metadata represents a set of media's metadata, it is essentially a set of
// (key, value).
type Metadata = map[string]interface{}

// File represents a media file
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

// ReadMetadata reads metadata from the provided File and setup the proper media
// Type if not done by the corresponding Handler.
func ReadMetadata(f File) (Metadata, error) {
	mh, err := handlers.ForReader(f)
	if err != nil {
		return nil, err
	}

	mdata, err := mh.ReadMetadata(f)
	if err != nil {
		return nil, err
	}

	mdata[TypeField] = mh.Type()

	return mdata, nil
}

// ReadMetadataFromFile reads metadata from the provided file name.
func ReadMetadataFromFile(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadMetadata(f)
}

// FetchMetadata retrieves the metadata from an external source (usually an
// internet data base) that corresponds to the provided known data.
func FetchMetadata(mdata Metadata) ([]Metadata, error) {
	mh, err := handlers.ForMetadata(mdata)
	if err != nil {
		return nil, err
	}

	return mh.FetchMetadata(mdata)
}

// CheckMetadata assesses the quality level of a set of metadata on a 0 to 100
// scale (0: very bad, 100: perfect). If the provided metadata is of unknown
// type, returns 0.
func CheckMetadata(mdata Metadata) int {
	mh, err := handlers.ForMetadata(mdata)
	if err != nil {
		return 0
	}

	return mh.CheckMetadata(mdata)
}

// IDCard returns an identity card  of a given media, notably to support
// query operation looking after this media. An identity is made of two
// parts, one that capture metadata that shall be unique, one that captures
// metadata that can be similar but pointing to the same media.
func IDCard(mdata Metadata) (exact [][2]string, similar [][2]string) {
	mh, err := handlers.ForMetadata(mdata)
	if err != nil {
		return
	}

	return mh.IDCard(mdata)
}
