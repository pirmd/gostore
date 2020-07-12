package books

import (
	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/media/books/googlebooks"
)

func fetchMetadata(mdata media.Metadata) ([]media.Metadata, error) {
	found, err := googlebooks.Search(mdata)
	if err != nil {
		return nil, err
	}

	return found, nil
}
