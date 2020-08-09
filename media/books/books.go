package books

import (
	"strconv"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/media/books/googlebooks"

	"github.com/pirmd/gostore/util"
)

type bookHandler struct{}

func (bh *bookHandler) FetchMetadata(mdata media.Metadata) ([]media.Metadata, error) {
	found, err := googlebooks.SearchVolume(mdata2vol(mdata))
	if err != nil {
		return nil, err
	}

	var res []media.Metadata
	for _, vi := range found {
		res = append(res, vol2mdata(vi))
	}

	return res, nil
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

func mdata2vol(mdata media.Metadata) *googlebooks.VolumeInfo {
	vi := &googlebooks.VolumeInfo{}

	if isbn, ok := mdata["ISBN"].(string); ok {
		vi.Identifier = append(vi.Identifier, googlebooks.Identifier{Type: "ISBN", Identifier: isbn})
	}

	if title, ok := mdata["Title"].(string); ok {
		vi.Title = title
	}

	if authors, ok := mdata["Authors"].([]string); ok {
		vi.Authors = append(vi.Authors, authors...)
	}

	if publisher, ok := mdata["Publisher"].(string); ok {
		vi.Publisher = publisher
	}

	return vi
}

func vol2mdata(vi *googlebooks.VolumeInfo) media.Metadata {
	mdata := make(media.Metadata)

	title, serie, seriePos := GuessSerie(vi.Title)
	subtitle := vi.SubTitle
	if len(serie) == 0 {
		subtitle, serie, seriePos = GuessSerie(vi.SubTitle)
	}

	mdata["Title"] = title

	if len(subtitle) > 0 {
		mdata["SubTitle"] = subtitle
	}

	if len(serie) > 0 {
		mdata["Serie"] = serie
		mdata["SeriePosition"] = seriePos
	}

	mdata["Authors"] = vi.Authors

	mdata["Description"] = vi.Description

	if len(vi.Subject) > 0 {
		mdata["Subject"] = vi.Subject
	}

	if len(vi.Publisher) > 0 {
		mdata["Publisher"] = vi.Publisher
	}

	if len(vi.PublishedDate) > 0 {
		if stamp, err := util.ParseTime(vi.PublishedDate); err != nil {
			mdata["PublishedDate"] = vi.PublishedDate
		} else {
			mdata["PublishedDate"] = stamp
		}
	}

	if vi.PageCount > 0 {
		mdata["PageCount"] = vi.PageCount
	}

	if len(vi.Language) > 0 {
		mdata["Language"] = vi.Language
	}

	for _, id := range vi.Identifier {
		if id.Type == "ISBN_13" && len(id.Identifier) > 0 {
			mdata["ISBN"] = id.Identifier
		}
	}

	return mdata
}
