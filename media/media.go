package media

import (
	"os"
    "fmt"
)

const (
    //TypeField is the name of the field that contains the type of the media
    //file
    TypeField = "Type"
)

var (
    //No book found correponding to the given metadata set
    ErrNoMatchFound = fmt.Errorf("No match found")
)

//GetMetadata reads metadata from the provided File and setup the proper media
//Type if not done by the corresponding Handler.
func GetMetadata(f File) (map[string]interface{}, error) {
	mh, err := handlers.ForReader(f)
	if err != nil {
		return nil, err
	}

	mdata, err := mh.GetMetadata(f)
	if err != nil {
		return nil, err
	}

	if _, exists := mdata[TypeField]; !exists {
		mdata[TypeField] = mh.Type()
	}

	return mdata, nil
}

//GetMetadataFromFile reads metadata from the provided filename
func GetMetadataFromFile(path string) (map[string]interface{}, error) {
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
func FetchMetadata(metadata map[string]interface{}) (map[string]interface{}, error) {
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

	if _, exists := mdata[TypeField]; mdata != nil && !exists {
		mdata[TypeField] = mh.Type()
	}

	return mdata, nil
}
