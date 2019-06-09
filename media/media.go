package media

import (
	"fmt"
	"io"
	"os"

	"github.com/pirmd/gostore/store"
)

const (
	//TypeField is the name of the field that contains the type of the media
	//file
	TypeField = "Type"
)

var (
	//ErrNoMetadataFound reports an error when no Metadata found
	ErrNoMetadataFound = fmt.Errorf("No metadata found")
)

//Metadata represents a set of media metadata, it is essentielly a set of (key,
//values). It is an alias to store.Value to benefit of its helpers functions
type Metadata = store.Value

//File represents a media file
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

//GetMetadata reads metadata from the provided File and setup the proper media
//Type if not done by the corresponding Handler.
func GetMetadata(f File) (Metadata, error) {
	mh, err := handlers.ForReader(f)
	if err != nil {
		return nil, err
	}

	mdata, err := mh.GetMetadata(f)
	if err != nil {
		return nil, err
	}

	mdata.SetIfNotExists(TypeField, mh.Type())

	return mdata, nil
}

//GetMetadataFromFile reads metadata from the provided filename
func GetMetadataFromFile(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return GetMetadata(f)
}

//FetchMetadata retrieves the metadata from an external source (usually an
//internet data base) that best correspond to the provided known data. It
//provides its best guess or nil if nothing reasonable is found.
//
//FetchMetadata uses as input any known metadata to based its search on and
//verifying that what it guesses is similar enough to its input to match.
//Provided input shall provide a valid media Type otherwise FetchMetadata will
//panic.
//
//FetchMetadata set the proper media Type if not done by the corresponding
//Handler.
func FetchMetadata(metadata Metadata) (Metadata, error) {
	typ, ok := metadata[TypeField].(string)
	if !ok {
		panic("metadata type is unknown or not of type 'string'")
	}

	mh, err := handlers.ForType(typ)
	if err != nil {
		return nil, err
	}

	mdata, err := mh.FetchMetadata(metadata)
	if err != nil {
		return nil, err
	}

	mdata.SetIfNotExists(TypeField, mh.Type())

	return mdata, nil
}
