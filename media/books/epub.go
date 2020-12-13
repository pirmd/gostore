package books

import (
	"archive/zip"
	"io"
	"path/filepath"

	"github.com/pirmd/epub"
	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/media/books/sanitizer"
	"github.com/pirmd/gostore/util"
)

var (
	_ media.Handler = (*epubHandler)(nil)
)

type epubHandler struct {
	*bookHandler
}

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

	mh.bookHandler.CleanMetadata(mdata)

	return mdata, nil
}

func (mh *epubHandler) Check(mdata media.Metadata, f media.File) (map[string]string, error) {
	findings, err := mh.bookHandler.Check(mdata, f)
	if err != nil {
		return nil, err
	}

	scanErr := new(util.MultiErrors)
	if err := mh.WalkContent(f, func(path string, r io.Reader, err error) error {
		if filepath.Ext(path) != ".html" {
			return nil
		}

		if err != nil {
			scanErr.Add(err)
			return nil
		}

		if err := sanitizer.EPUB.Scan(r); err != nil {
			scanErr.Add(err)
		}

		return nil
	}); err != nil {
		scanErr.Add(err)
	}

	if scanErr.Err() != nil {
		findings["Content"] = scanErr.Error()
	}

	return findings, nil
}

func (mh *epubHandler) WalkContent(f media.File, walkFn media.WalkFunc) error {
	sz, err := getSize(f)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(f, sz)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		rf, err := f.Open()
		if err == nil {
			defer rf.Close()
		}

		if err := walkFn(f.Name, rf, err); err != nil {
			return err
		}
	}

	return nil
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
		}
	}

	if len(epubData.Publisher) > 0 {
		mdata["Publisher"] = epubData.Publisher[0]
	}

	for _, d := range epubData.Date {
		if d.Event == "publication" {
			mdata["PublishedDate"] = d.Stamp
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

func getSize(f io.Seeker) (int64, error) {
	sz, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	return sz, nil
}

func init() {
	media.RegisterHandler(&epubHandler{})
}
