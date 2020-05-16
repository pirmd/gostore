package books

import (
	"github.com/pirmd/epub"
	"github.com/pirmd/gostore/media"
)

var (
	_ media.Handler = (*epubHandler)(nil)
)

//TODO: rename ISBN to ISBN_13 (?)

type epubHandler struct{}

func (mh *epubHandler) Type() string {
	return "book/epub"
}

func (mh *epubHandler) Mimetype() string {
	return "application/epub+zip"
}

func (mh *epubHandler) GetMetadata(f media.File) (media.Metadata, error) {
	epubData, err := epub.GetMetadata(f)
	if err != nil {
		return nil, err
	}
	return epub2mdata(epubData), nil
}

func epub2mdata(epubData *epub.Metadata) media.Metadata {
	mdata := make(media.Metadata)

	if len(epubData.Title) > 0 {
		mdata.Set("Title", epubData.Title[0])
	}

	if len(epubData.Creator) > 0 {
		authors := []string{}
		for _, c := range epubData.Creator {
			authors = append(authors, c.FullName)
		}
		mdata.Set("Authors", authors)
	}

	if len(epubData.Description) > 0 {
		mdata.Set("Description", epubData.Description[0])
	}

	if len(epubData.Subject) > 0 {
		mdata.Set("Subject", epubData.Subject)
	}

	for _, id := range epubData.Identifier {
		if id.ID == "isbn" {
			mdata.Set("ISBN", id.Value)
			break
		}
	}

	if len(epubData.Publisher) > 0 {
		mdata.Set("Publisher", epubData.Publisher[0])
	}

	for _, d := range epubData.Date {
		if d.Event == "publication" {
			mdata.Set("PublishedDate", d.Stamp)
		}
	}

	for _, meta := range epubData.Meta {
		switch meta.Name {
		case "calibre:series":
			mdata.Set("Serie", meta.Content)

		case "calibre:series_index":
			mdata.Set("SeriePosition", meta.Content)
		}
	}

	return mdata
}

func init() {
	media.RegisterHandler(&epubHandler{})
}
