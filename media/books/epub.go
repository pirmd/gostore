package books

import (
	"archive/zip"
	"io"

	"github.com/pirmd/epub"
	"github.com/pirmd/gostore/media"
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

func (mh *epubHandler) ProcessContent(w io.Writer, f media.File, procFn media.ProcessingFunc, filters ...func(string) bool) error {
	sz, err := getSize(f)
	if err != nil {
		return err
	}

	r, err := zip.NewReader(f, sz)
	if err != nil {
		return err
	}

	zipw := zip.NewWriter(w)
	defer func() {
		cerr := zipw.Close()
		if err == nil {
			err = cerr
		}
	}()

	procErr := &util.MultiErrors{}

NextFile:
	for _, f := range r.File {
		rf, err := f.Open()
		if err != nil {
			procErr.Add(err)
			continue
		}
		defer rf.Close()

		header, err := zip.FileInfoHeader(f.FileInfo())
		if err != nil {
			procErr.Add(err)
			continue
		}
		header.Name = f.Name
		wf, err := zipw.CreateHeader(header)
		if err != nil {
			procErr.Add(err)
			continue
		}

		for _, filter := range filters {
			if filter(f.Name) {
				_, err = io.Copy(wf, rf)
				if err != nil {
					procErr.Add(err)
				}
				continue NextFile
			}
		}

		if err := procFn(wf, rf); err != nil {
			procErr.Add(err)
		}
	}

	return procErr.Err()
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
