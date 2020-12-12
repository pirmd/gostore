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

// Check reviews a media metadata and content to capture possible quality issues.
// Quality issues are organised by Metadata Field using "Content" as the
// special field name for issues about the media.File content itself.
func Check(mdata Metadata, f File) (map[string]string, error) {
	mh, err := handlers.ForMetadata(mdata)
	if err != nil {
		return nil, err
	}

	return mh.Check(mdata, f)
}

// ProcessContent processes a media File by applying a processing function
// to its content and feedbacking the result to the given io.Writer.
// It accepts a set of filters should the processing only apply to a subset
// of the media.File content, should media be made of several parts (like a
// zip archive).
func ProcessContent(w io.Writer, f File, procFn ProcessingFunc, filters ...func(string) bool) error {
	mh, err := handlers.ForReader(f)
	if err != nil {
		return err
	}

	return mh.ProcessContent(w, f, procFn, filters...)
}
