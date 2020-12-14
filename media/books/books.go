package books

import (
	"strconv"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/media/books/googlebooks"

	"github.com/pirmd/gostore/util"
)

// bookHandler offers generic functions helpful for any e-book handlers.
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

	for _, mdata := range res {
		bh.CleanMetadata(mdata)
	}

	return res, nil
}

func (bh *bookHandler) Check(mdata media.Metadata, f media.File) (map[string]string, error) {
	findings := make(map[string]string)

	if util.IsZero(mdata["Title"]) {
		findings["Title"] = "missing"
	}

	if util.IsZero(mdata["Authors"]) {
		findings["Authors"] = "missing"
	}

	if util.IsZero(mdata["Description"]) {
		findings["Description"] = "missing"
	}

	if util.IsZero(mdata["Publisher"]) {
		findings["Publisher"] = "missing"
	}

	if util.IsZero(mdata["Serie"]) && (util.IsNotZero(mdata["SeriePosition"]) || util.IsNotZero(mdata["SerieEpisode"])) {
		findings["Serie"] = "incomplete serie information"
	}
	if util.IsZero(mdata["SeriePosition"]) && (util.IsNotZero(mdata["Serie"]) || util.IsNotZero(mdata["SerieEpisode"])) {
		findings["SeriePosition"] = "incomplete serie information"
	}
	if util.IsZero(mdata["SerieEpisode"]) && (util.IsNotZero(mdata["Serie"]) || util.IsNotZero(mdata["SeriePosition"])) {
		findings["SerieEpisode"] = "incomplete serie information"
	}

	return findings, nil
}

func (bh *bookHandler) CleanMetadata(mdata media.Metadata) {
	episode, serie, seriePos := mdata["SerieEpisode"], mdata["Serie"], mdata["SeriePosition"]

	if _, exist := mdata["Title"]; exist {
		e, s, p := GuessSerie(mdata["Title"].(string))
		episode, serie, seriePos = firstNonZero(episode, e), firstNonZero(serie, s), firstNonZero(seriePos, p)
	}

	if _, exist := mdata["SubTitle"]; exist {
		e, s, p := GuessSerie(mdata["SubTitle"].(string))
		episode, serie, seriePos = firstNonZero(episode, e), firstNonZero(serie, s), firstNonZero(seriePos, p)
	}

	if !util.IsZero(serie) {
		mdata["Serie"] = serie

		mdata["SeriePosition"] = seriePos

		if pos, ok := seriePos.(string); ok {
			if nb, err := strconv.Atoi(pos); err == nil {
				mdata["SeriePosition"] = nb
			}
		}

		mdata["SerieEpisode"] = episode
	}

	if date, exist := mdata["PublishedDate"]; exist {
		if d, ok := date.(string); ok {
			if stamp, err := util.ParseTime(d); err == nil {
				mdata["PublishedDate"] = stamp
			}
		}
	}

	if pages, exist := mdata["PageCount"]; exist {
		if p, ok := pages.(string); ok {
			if nb, err := strconv.Atoi(p); err == nil {
				mdata["PageCount"] = nb
			}
		}
	}
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

	if len(vi.Title) > 0 {
		mdata["Title"] = vi.Title
	}

	if len(vi.SubTitle) > 0 {
		mdata["SubTitle"] = vi.SubTitle
	}

	if len(vi.Authors) > 0 {
		mdata["Authors"] = vi.Authors
	}

	if len(vi.Description) > 0 {
		mdata["Description"] = vi.Description
	}

	if len(vi.Subject) > 0 {
		mdata["Subject"] = vi.Subject
	}

	if len(vi.Publisher) > 0 {
		mdata["Publisher"] = vi.Publisher
	}

	if len(vi.PublishedDate) > 0 {
		mdata["PublishedDate"] = vi.PublishedDate
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

func firstNonZero(a, b interface{}) interface{} {
	if util.IsZero(a) {
		return b
	}
	return a
}
