package books

import (
	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/media/books/googlebooks"

	"github.com/pirmd/gostore/util"
)

func fetchMetadata(mdata media.Metadata) ([]media.Metadata, error) {
	found, err := googlebooks.Search(mdata)
	if err != nil {
		return nil, err
	}

	return found, nil
}

func checkMetadata(mdata media.Metadata) int {
	lvl := 100

	if util.IsZero(mdata["Title"]) {
		lvl = 0
	}

	if util.IsZero(mdata["Authors"]) {
		lvl -= 50
	}

	if util.IsZero(mdata["Description"]) {
		lvl -= 20
	}

	if util.IsZero(mdata["Publisher"]) {
		lvl -= 10
	}

	if lvl < 0 {
		lvl = 0
	}

	return lvl
}
