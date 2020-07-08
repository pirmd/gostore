package googlebooks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/pirmd/gostore/media"
	"github.com/pirmd/gostore/util"
)

const (
	googleBooksURL = "https://www.googleapis.com/books/v1/volumes"
)

// TODO(pirmd): check again Google book API to improve the query

var (
	// reSerieGuesser is a collection of regexp to extract series information
	// from title/subtitles.
	// It should be made of 3 named capturing groups (title, serie, serie number).
	reSerieGuesser = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<title>.+)\s\((?P<serie>.+?)\s(?i:#|Series |n°|)(?P<seriePos>\d+)\)$`),
		regexp.MustCompile(`^(?P<title>.+)\s-\s(?P<serie>.+?)\s(?i:#|Series |n°|)(?P<seriePos>\d+)$`),
		regexp.MustCompile(`^(?P<serie>.+?)\s(?i:#|Series |n°|)(?P<seriePos>\d+)$`),
		regexp.MustCompile(`^Book\s(?P<seriePos>\d+)\sof\s(?P<serie>.+)$`),
	}
)

type googleVolumes struct {
	Items []*googleVolume `json:"items"`
}

type googleVolume struct {
	VolumeInfo *googleVolumeInfo `json:"volumeInfo"`
}

type googleVolumeInfo struct {
	Title         string       `json:"title"`
	SubTitle      string       `json:"subtitle"`
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

// googleBooks wraps Google books API
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
	if title, ok := mdata["Title"].(string); ok {
		query = append(query, "intitle:"+title)
	}

	if authors, ok := mdata["Authors"].([]string); ok {
		query = append(query, "inauthor:"+strings.Join(authors, "+"))
	}

	if isbn, ok := mdata["ISBN"].(string); ok {
		query = append(query, "isbn:"+isbn)
	}

	if len(query) == 0 {
		return "", fmt.Errorf("empty query")
	}

	q := url.Values{}
	q.Set("q", strings.Join(query, "+"))
	q.Set("orderBy", "relevance")
	q.Set("printType", "books")
	q.Set("maxResults", "5")

	return googleBooksURL + "?" + q.Encode(), nil
}

func (g *googleBooks) vol2mdata(vi *googleVolumeInfo) media.Metadata {
	mdata := make(media.Metadata)
	title, subtitle, serie, seriePos := g.parseTitle(vi)

	mdata["Title"] = title
	mdata["Authors"] = vi.Authors
	mdata["Description"] = vi.Description

	if len(vi.Subject) > 0 {
		mdata["Subject"] = vi.Subject
	}

	if len(subtitle) > 0 {
		mdata["SubTitle"] = subtitle
	}

	if len(serie) > 0 {
		mdata["Serie"] = serie
		mdata["SeriePosition"] = seriePos
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

// parseTitle use a simple heuristic to decipher Google books information about
// series hidden in title/subtitle volume information
func (g *googleBooks) parseTitle(vi *googleVolumeInfo) (title string, subtitle string, serieName string, seriePos string) {
	for _, re := range reSerieGuesser {
		if r := submatchMap(re, vi.Title); len(r) > 0 {
			return r["title"], vi.SubTitle, r["serie"], r["seriePos"]
		}

		if r := submatchMap(re, vi.SubTitle); len(r) > 0 {
			return vi.Title, r["title"], r["serie"], r["seriePos"]
		}
	}

	return vi.Title, vi.SubTitle, "", ""
}

func submatchMap(re *regexp.Regexp, s string) map[string]string {
	names := re.SubexpNames()
	matches := re.FindStringSubmatch(s)

	r := make(map[string]string)
	for i := range matches {
		if i > 0 {
			r[names[i]] = matches[i]
		}
	}

	return r
}
