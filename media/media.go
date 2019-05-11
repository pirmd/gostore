package media

import (
	"os"
)

//TypeField is the name of MediaFile.mdata key that contains
//the type of the media file
const TypeField = "Type"

//GetMetadata reads metadata from the provided File and setup the proper media Type.
//Reading metadata and setting the media type is done through the corresponfing
//MediaHandlers
func GetMetadata(f File) (map[string]interface{}, error) {
	mh, err := handlers.For(f)
	if err != nil {
		return nil, err
	}

	mdata, err := mh.GetMetadata(f)
	if err != nil {
		return nil, err
	}

	if _, exists := mdata[TypeField]; !exists {
		mdata[TypeField] = mh.Name()
	}

	if _, ok := mdata[TypeField].(string); !ok {
		panic("metadata type is not of type 'string'")
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
