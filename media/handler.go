package media

import (
	"fmt"
	"io"

	"github.com/gabriel-vasile/mimetype"
)

const (
    //DefaultMimetype is the fallback content type to identify file.
    //The handler that supports DefaultMimetype, if any, is also the default
    //handler should no handler be registered for a given Type or Mimetype.
    DefaultMimetype = "application/octet-stream"
)

var (
    //ErrUnknownMediaType error is raise if submitted file is of unknown media
    //type, that is to say that no corresponding media handler is found
	ErrUnknownMediaType = fmt.Errorf("media handler: unknown file type")

	handlers = Handlers{} //register of all media handlers known
)

//Handler represents a media handler
type Handler interface {
    //Type provides the name of the handler. An handler's name is mainly used to
    //identify a media type and adopt customized behavior based on a media type.
	Type() string

    //Mimetype provides the mimetype that the handler can manage.
    Mimetype() string

    //GetMetadata retrieves the metadata from a given file.
	GetMetadata(File) (map[string]interface{}, error)

    //FetchMetadata retrieves the metadata from an external source (usually
    //an internet data base) that best correspond to the provided known data.
    //It provides its best guess or nil if nothing reasonable is found.
    FetchMetadata(map[string]interface{}) (map[string]interface{}, error)
}

type Handlers []Handler

//ForReader retrieves the handler for corresponding to the given file based on
//the file guessed content type Should no registered handler is found, the
//handler for DefaultType is returned if it exists.
//
//ErrUnknownMediaType is returned if no handler is found.
func (h Handlers) ForReader(f io.Reader) (Handler, error) {
	mtyp, err := getMimetype(f)
	if err != nil {
		return nil, err
	}

    for _, mh := range h {
        if mh.Mimetype() == mtyp {
            return mh, nil
        }
    }

	return h.defaultHandler()
}

//ForType retrieves the handler corresponding to the provided type.
//Should no registered handler is found, the handler for DefaultType is
//returned if it exists.
//
//ErrUnknownMediaType is returned if no handler is found.
func (h Handlers) ForType(typ string) (Handler, error) {
    for _, mh := range h {
        if mh.Type() == typ {
            return mh, nil
        }
    }

	return h.defaultHandler()
}

//defaultHandler is the handler that manages DefaultMimetype
func (h Handlers) defaultHandler() (Handler, error) {
    for _, mh := range h {
        if mh.Mimetype() == DefaultMimetype {
            return mh, nil
        }
    }

	return nil, ErrUnknownMediaType
}

//RegisterHandler registers a new media handler
//It will panic if a media handler with the same Type() or Mimetype() already
//exists
func RegisterHandler(mh Handler) {
    for _, h := range handlers {
        if mh.Type() == h.Type() || mh.Mimetype() == h.Mimetype() {
            panic("a media handler already exists for " + mh.Type() + ":" + mh.Mimetype())
        }
    }

	handlers = append(handlers, mh)
}

// Always returns a valid content-type by returning "application/octet-stream" if no others seemed to match.
func getMimetype(r io.Reader) (mime string, err error) {
	mime, _, err = mimetype.DetectReader(r)
	return
}

//File represents a media file
type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}
