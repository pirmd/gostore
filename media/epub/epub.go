package epub

import (
	"strconv"

	"github.com/pirmd/epub"
	"github.com/pirmd/gostore/media"
)

type epubHandler struct{}

func (mh *epubHandler) Name() string {
	return "epub"
}

func (mh *epubHandler) GetMetadata(f media.File) (map[string]interface{}, error) {
	mdata, err := epub.GetMetadata(f)
	if err != nil {
		return nil, err
	}
	return epub2map(mdata), nil
}

func epub2map(mdata *epub.Metadata) map[string]interface{} {
	m := make(map[string]interface{})

	if len(mdata.Title) > 0 {
		m["Title"] = mdata.Title[0]
	}

	if len(mdata.Creator) > 0 {
		authors := []string{}
		for _, c := range mdata.Creator {
			authors = append(authors, c.FullName)
		}
		m["Authors"] = authors
	}

	if len(mdata.Description) > 0 {
		m["Description"] = mdata.Description[0]
	}

	m["Subject"] = mdata.Subject

	for _, id := range mdata.Identifier {
		if id.ID == "isbn" {
			m["ISBN"] = id.Value
			break
		}
	}

	if len(mdata.Publisher) > 0 {
		m["Publisher"] = mdata.Publisher[0]
	}

	for _, d := range mdata.Date {
		if d.Event == "publication" {
			if t, err := parseTime(d.Stamp); err == nil {
				m["PublicationDate"] = t
			}
			break
		}
	}

	for _, meta := range mdata.Meta {
		switch meta.Name {
		case "calibre:series":
			m["Serie"] = meta.Content

		case "calibre:series_index":
			if pos, err := strconv.ParseFloat(meta.Content, 32); err == nil {
				m["SeriePosition"] = pos
			}
		}
	}

	return m
}

func init() {
	media.RegisterHandler("application/epub+zip", &epubHandler{})
}
