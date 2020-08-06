package books

import (
	"github.com/pirmd/gostore/util"

	"github.com/pirmd/epub"
	"github.com/pirmd/gostore/media"
)

var (
	_ media.Handler = (*epubHandler)(nil)
)

type epubHandler struct{}

func (mh *epubHandler) Type() string {
	return "book/epub"
}

func (mh *epubHandler) Mimetype() string {
	return "application/epub+zip"
}

func (mh *epubHandler) ReadMetadata(f media.File) (media.Metadata, error) {
	epubData, err := epub.GetMetadata(f)
	if err != nil {
		return nil, err
	}

	mdata := epub2mdata(epubData)
	mdata[media.TypeField] = mh.Type()
	return mdata, nil
}

func (mh *epubHandler) FetchMetadata(mdata media.Metadata) ([]media.Metadata, error) {
	metadata, err := fetchMetadata(mdata)
	if err != nil {
		return nil, err
	}

	for _, m := range metadata {
		m[media.TypeField] = mh.Type()
	}

	return metadata, nil
}

func (mh *epubHandler) CheckMetadata(mdata media.Metadata) int {
	return checkMetadata(mdata)
}

func (mh *epubHandler) IDCard(mdata media.Metadata) (exact [][2]string, similar [][2]string) {
	return identity(mdata)
}

func epub2mdata(epubData *epub.Metadata) media.Metadata {
	mdata := make(media.Metadata)

	if len(epubData.Title) > 0 {
		mdata["Title"] = epubData.Title[0]
	}

	if len(epubData.Creator) > 0 {
		authors := []string{}
		for _, c := range epubData.Creator {
			authors = append(authors, c.FullName)
		}
		mdata["Authors"] = authors
	}

	if len(epubData.Description) > 0 {
		mdata["Description"] = epubData.Description[0]
	}

	if len(epubData.Subject) > 0 {
		mdata["Subject"] = epubData.Subject
	}

	for _, id := range epubData.Identifier {
		if id.ID == "isbn" {
			mdata["ISBN"] = id.Value
			break
		}
	}

	if len(epubData.Publisher) > 0 {
		mdata["Publisher"] = epubData.Publisher[0]
	}

	for _, d := range epubData.Date {
		if d.Event == "publication" {
			if stamp, err := util.ParseTime(d.Stamp); err != nil {
				mdata["PublishedDate"] = d.Stamp
			} else {
				mdata["PublishedDate"] = stamp
			}
		}
	}

	for _, meta := range epubData.Meta {
		switch meta.Name {
		case "calibre:series":
			mdata["Serie"] = meta.Content

		case "calibre:series_index":
			mdata["SeriePosition"] = meta.Content
		}
	}

	return mdata
}

func init() {
	media.RegisterHandler(&epubHandler{})
}
