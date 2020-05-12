package media

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pirmd/gostore/store"
)

const (
	// TypeField is the name of the field that contains the type of the media
	TypeField = "Type"

	// DefaultMediaType is the name used by default to identify a media type.
	// It is usually used in case media cannot be identified (either missing
	// or incorrect value)
	DefaultMediaType = "media"
)

var (
	// ErrNoMetadataFound reports an error when no Metadata found
	ErrNoMetadataFound = fmt.Errorf("no metadata found")
)

// Metadata represents a set of media's metadata, it is essentially a set of (key,
// value).
type Metadata = store.Value

// Type returns the media's type corresponding to the TypeField attribute
// value. If no information exists for TypeField attribute or if it is not of
// the appropriate type, it feedbacks DefaultMediaType
func Type(mdata Metadata) string {
	typ, ok := mdata.Get(TypeField).(string)
	if !ok || typ == "" {
		return DefaultMediaType
	}

	return typ
}

// IsOfType checks whether the media is of the given type.
// IsOfType considers media's type of the form "family/sub-family" and checks
// if the provided type name is either the complete type, only the family of the
// sub-family.
func IsOfType(mdata Metadata, typ string) bool {
	t := Type(mdata)
	return t == typ || strings.HasPrefix(t, typ) || strings.HasSuffix(t, typ)
}

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
