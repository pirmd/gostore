package media

import (
	"strings"
)

const (
	// TypeField is the name of the field that contains the type of the media
	TypeField = "Type"

	// DefaultType is the name used by default to identify a media type.
	// It is usually used in case media cannot be identified (either missing
	// or incorrect value)
	DefaultType = "media"
)

// TypeOf returns the media's type corresponding to the TypeField attribute
// value. If no information exists for TypeField attribute or if it is not of
// the appropriate type, it feedbacks DefaultType
func TypeOf(mdata Metadata) string {
	typ, ok := mdata[TypeField].(string)
	if !ok || typ == "" {
		return DefaultType
	}

	return typ
}

// IsOfType checks whether the media is of the given type.
// IsOfType considers media's type of the form "family/sub-family" and checks
// if the provided type name is either the complete type, only the family of the
// sub-family.
func IsOfType(mdata Metadata, typ string) bool {
	t := TypeOf(mdata)
	return t == typ || strings.HasPrefix(t, typ+"/") || strings.HasSuffix(t, "/"+typ)
}
