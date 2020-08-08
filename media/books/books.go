package books

import (
	"strconv"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/media/books/googlebooks"

	"github.com/pirmd/gostore/util"
)

type bookHandler struct{}

func (bh *bookHandler) FetchMetadata(mdata media.Metadata) ([]media.Metadata, error) {
	found, err := googlebooks.Search(mdata)
	if err != nil {
		return nil, err
	}

	return found, nil
}

func (bh *bookHandler) CheckMetadata(mdata media.Metadata) int {
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

func (bh *bookHandler) IDCard(mdata media.Metadata) (exact [][2]string, similar [][2]string) {
	if isbn, ok := mdata["ISBN"].(string); ok {
		exact = append(exact, [2]string{"ISBN", isbn})
	}

	if title, ok := mdata["Title"].(string); ok {
		similar = append(similar, [2]string{"Title", title})
	}
	if serie, ok := mdata["Serie"].(string); ok {
		similar = append(similar, [2]string{"Serie", serie})
	}
	if seriePosition, ok := mdata["SeriePosition"].(int); ok {
		similar = append(similar, [2]string{"SeriePosition", strconv.Itoa(seriePosition)})
	}
	if authors, ok := mdata["Authors"].([]string); ok {
		for _, a := range authors {
			similar = append(similar, [2]string{"Authors", a})
		}
	}
	if publisher, ok := mdata["Publisher"].(string); ok {
		similar = append(similar, [2]string{"Publisher", publisher})
	}

	return
}
