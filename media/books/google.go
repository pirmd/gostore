package books

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pirmd/gostore/media"
)

const googleBooksURL = "https://www.googleapis.com/books/v1/volumes"

type googleVolumes struct {
	Items []*googleVolume `json:"items"`
}

type googleVolume struct {
	VolumeInfo *googleVolumeInfo `json:"volumeInfo"`
}

type googleVolumeInfo struct {
	Title         string       `json:"title"`
	Language      string       `json:"language"`
	Identifier    []identifier `json:"industryIdentifiers"`
	Authors       []string     `json:"authors"`
	Subject       []string     `json:"categories"`
	Description   string       `json:"description"`
	Publisher     string       `json:"publisher"`
	PublishedDate string       `json:"publishedDate"`
	PageCount     int64        `json:"pageCount"`
}

type identifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

//googleBooks wraps google books api into a Fetcher
type googleBooks struct{}

func (g *googleBooks) LookForBooks(mdata media.Metadata) ([]media.Metadata, error) {
	queryURL, err := g.buildQueryURL(mdata)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, media.ErrNoMetadataFound
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vol *googleVolumes
	if err := json.Unmarshal(data, &vol); err != nil {
		return nil, err
	}

	var metadata []media.Metadata
	for _, v := range vol.Items {
		metadata = append(metadata, g.vol2mdata(v.VolumeInfo))
	}

	return metadata, nil
}

func (g *googleBooks) buildQueryURL(mdata media.Metadata) (string, error) {
	var query []string
	if title, ok := mdata.Get("Title").(string); ok {
		query = append(query, "intitle:"+title)
	}

	if authors, ok := mdata.Get("Authors").([]string); ok {
		query = append(query, "inauthor:"+strings.Join(authors, "+"))
	}

	if isbn, ok := mdata.Get("ISBN").(string); ok {
		query = append(query, "isbn:"+isbn)
	}

	if len(query) == 0 {
		return "", fmt.Errorf("Empty query")
	}

	q := url.Values{}
	q.Set("q", strings.Join(query, "+"))
	q.Set("orderBy", "relevance")
	q.Set("printType", "books")
	q.Set("maxResults", "5") //TODO: config this part
	//TODO: q.Set("langRestrict", "fr") Is it realy needed?

	return googleBooksURL + "?" + q.Encode(), nil
}

func (g *googleBooks) vol2mdata(vi *googleVolumeInfo) media.Metadata {
	mdata := make(media.Metadata)
	mdata.Set("Title", vi.Title)
	mdata.Set("Authors", vi.Authors)
	mdata.Set("Description", vi.Description)
	mdata.Set("Subject", vi.Subject)
	mdata.Set("Publisher", vi.Publisher)
	mdata.Set("PublishedDate", vi.PublishedDate)
	mdata.Set("PageCount", vi.PageCount)
	mdata.Set("Language", vi.Language)

	for _, id := range vi.Identifier {
		if id.Type == "ISBN_13" {
			mdata.Set("ISBN", id.Identifier)
		}
	}

	return mdata
}
