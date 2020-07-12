package media

import (
	"fmt"
	"io"

	"github.com/gabriel-vasile/mimetype"
)

const (
	// DefaultMimetype is the fallback content type to identify file. The
	// handler that supports DefaultMimetype, if any, is also the default
	// handler should no handler be registered for a given Type or Mimetype.
	DefaultMimetype = "application/octet-stream"
)

var (
	// ErrUnknownMediaType error is raise if submitted file is of unknown media
	// type, that is to say that no corresponding media handler is found.
	ErrUnknownMediaType = fmt.Errorf("media handler: unknown file type")

	handlers = Handlers{} //register of all media known handlers
)

// Handler represents a media handler.
type Handler interface {
	// Type provides the name of the handler. An handler's name is mainly used
	// to identify a media type and adopt customized behavior based on a media
	// type.
	Type() string

	// Mimetype provides the mimetype that the handler can manage.
	Mimetype() string

	// GetMetadata retrieves the metadata from a given file.
	GetMetadata(File) (Metadata, error)

	// FetchMetadata retrieves the metadata from an external source (usually an
	// internet data base) that corresponds to the provided known data.
	FetchMetadata(Metadata) ([]Metadata, error)
}

// Handlers represent the list of known media handlers.
type Handlers []Handler

// ForReader retrieves the handler based on the reader guessed content type.
// Should no registered handler be found, the handler for DefaultType is
// returned if it exists.
//
// ErrUnknownMediaType is returned if no handler is found.
func (h Handlers) ForReader(f io.Reader) (Handler, error) {
	mtype, err := getMimetype(f)
	if err != nil {
		return nil, err
	}

	return h.ForMimetype(mtype)
}

// ForMimetype retrieves the handler corresponding to the provided mimetype.
// Should no registered handler be found, the handler for DefaultType is
// returned if it exists.
//
// ErrUnknownMediaType is returned if no handler is found.
func (h Handlers) ForMimetype(mtype string) (Handler, error) {
	for _, mh := range h {
		if mh.Mimetype() == mtype {
			return mh, nil
		}
	}

	return h.defaultHandler()
}

// ForType retrieves the handler corresponding to the provided type.  Should no
// registered handler be found, the handler for DefaultType is returned if it
// exists.
//
// ErrUnknownMediaType is returned if no handler is found.
func (h Handlers) ForType(typ string) (Handler, error) {
	for _, mh := range h {
		if mh.Type() == typ {
			return mh, nil
		}
	}

	return h.defaultHandler()
}

// defaultHandler is the handler that manages DefaultMimetype
func (h Handlers) defaultHandler() (Handler, error) {
	for _, mh := range h {
		if mh.Mimetype() == DefaultMimetype {
			return mh, nil
		}
	}

	return nil, ErrUnknownMediaType
}

// RegisterHandler registers a new media handler. It does not check if the
// handler is already registered.
func RegisterHandler(mh Handler) {
	handlers = append(handlers, mh)
}

// getMimetype returns mimetype for the provided reader.
// Always returns a valid content-type by returning "application/octet-stream"
// if no others type matched or if an error occurred.
func getMimetype(r io.Reader) (string, error) {
	mime, err := mimetype.DetectReader(r)
	return mime.String(), err
}
