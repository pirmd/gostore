package media

import (
	"fmt"
	"io"

	"github.com/gabriel-vasile/mimetype"
)

//DefaultType is the fallback content type to identify file
const DefaultType = "application/octet-stream"

var (
	//ErrUnknownMediaType error is raise if submitted file is of
	//unknown media type, that is to say that no corresponding
	//media handler is found
	ErrUnknownMediaType = fmt.Errorf("media handler: unknown file type")

	handlers = Handlers{} //register of all media handlers known
)

//Handler represents a media handler
type Handler interface {
	//Name provides the name of teh handler
	//An handler's name is mainly used to identify a media type and
	//adopt customize behavior based on media type (like specific naming
	//or printing scheme)
	Name() string

	//GetMetadata retrieves the metadata from a given file
	//Metadata outputs is free, specific key MediaTypeKey can be
	//used to enforce a media type
	GetMetadata(File) (map[string]interface{}, error)
}

type Handlers map[string]Handler

//For retrieves the handler for corresponding to the
//given file based on the file guessed content type
//Should no registered handler is found, the handler
//for DefaultType is returned if it exists.
//
//ErrUnknownMediaType is returned if no handler is found.
func (h Handlers) For(f io.Reader) (Handler, error) {
	typ, err := getType(f)
	if err != nil {
		return nil, err
	}

	if mh, exists := h[typ]; exists {
		return mh, nil
	}

	if mh, exists := h[DefaultType]; exists {
		return mh, nil
	}

	return nil, ErrUnknownMediaType
}

//RegisterHandler registers a new media handler
//It will panic if a media handler is added twice
func RegisterHandler(mimetype string, mh Handler) {
	if _, exists := handlers[mimetype]; exists {
		panic("mediahandler already exists for " + mimetype)
	}
	handlers[mimetype] = mh
}

// Always returns a valid content-type by returning "application/octet-stream" if no others seemed to match.
func getType(r io.Reader) (mime string, err error) {
	mime, _, err = mimetype.DetectReader(r)
	return
}

type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}
