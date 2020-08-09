package googlebooks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// API is documented in:
// .https://developers.google.com/books/docs/v1/reference/volumes
// .https://developers.google.com/books/docs/v1/using

const (
	// URL is the GoogleBooks API base URL used by this module.
	URL = "https://www.googleapis.com/books/v1/volumes"
)

var (
	// defaultAPI is the default  GoogleBooks API
	defaultAPI = &API{}
)

// API represents a GoogleBooks api.
type API struct {
	// OrderBy defines the query result sorting order. Accept:
	// . relevance - Returns results in order of the relevance of search terms
	// (default).
	// . newest - Returns results in order of most recently to least recently published.
	OrderBy string
	// MaxResults defines the maximum number of results to return. The default
	// is 10, and the maximum allowable value is 40.
	MaxResults int
}

// SearchVolume queries GoogleBooks API for books that corresponds to the
// provided VolumeInfo.
func (api *API) SearchVolume(vi *VolumeInfo) ([]*VolumeInfo, error) {
	queryURL := api.buildQueryURL(vi)
	if len(queryURL) == 0 {
		return nil, nil
	}

	resp, err := http.Get(queryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("googlebooks: query failed with status code %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vol *volumes
	if err := json.Unmarshal(data, &vol); err != nil {
		return nil, err
	}

	var res []*VolumeInfo
	for _, v := range vol.Items {
		res = append(res, v.VolumeInfo)
	}

	return res, nil
}

func (api *API) buildQueryURL(vi *VolumeInfo) string {
	query := vi.toQuery()
	if len(query) == 0 {
		return ""
	}

	q := url.Values{}
	q.Set("q", strings.Join(query, "+"))

	q.Set("printType", "books")

	if len(api.OrderBy) > 0 {
		q.Set("orderBy", "relevance")
	}

	if api.MaxResults > 0 {
		q.Set("maxResults", strconv.Itoa(api.MaxResults))
	}

	return URL + "?" + q.Encode()
}

// SearchVolume queries GoogleBooks API with some default parameters.
func SearchVolume(vi *VolumeInfo) ([]*VolumeInfo, error) {
	return defaultAPI.SearchVolume(vi)
}

type volumes struct {
	Items []*volume `json:"items"`
}

type volume struct {
	VolumeInfo *VolumeInfo `json:"volumeInfo"`
}

// Identifier represents an industry standard identifier.
type Identifier struct {
	// Type is the identifier type such as ISBN, ISBN_10, ISBN_13.
	Type string `json:"type"`
	// Identifier is the identifier value.
	Identifier string `json:"identifier"`
}

// VolumeInfo gathers information obtained from GoogleBooks API
type VolumeInfo struct {
	// Title is the volume's title.
	Title string `json:"title"`

	// SubTitle is the volume's sub-title.
	SubTitle string `json:"subtitle"`

	// Language is the volume's language. It is the two-letter ISO 639-1 code
	// such as 'fr', 'en'.
	Language string `json:"language"`

	// Identifier is the industry standard identifiers for this volume such as
	// ISBN_10, ISBN_13.
	Identifier []Identifier `json:"industryIdentifiers"`

	// Authors is the list names of the authors and/or editors for this volume.
	Authors []string `json:"authors"`

	// Subject is the list of subject categories, such as "Fiction",
	// "Suspense".
	Subject []string `json:"categories"`

	// Description is the synopsis of the volume. The text of the description
	// is formatted in HTML and includes simple formatting elements.
	Description string `json:"description"`

	// Publisher is the publisher of this volume.
	Publisher string `json:"publisher"`

	// PublishedDate is date of publication of this volume.
	PublishedDate string `json:"publishedDate"`

	// PageCount is total number of pages of this volume.
	PageCount int64 `json:"pageCount"`
}

func (vi *VolumeInfo) toQuery() (query []string) {
	if len(vi.Title) > 0 {
		query = append(query, "intitle:"+vi.Title)
	}

	if len(vi.Authors) > 0 {
		query = append(query, "inauthor:"+strings.Join(vi.Authors, "+"))
	}

	if len(vi.Publisher) > 0 {
		query = append(query, "intitle:"+vi.Title)
	}

	for _, id := range vi.Identifier {
		if strings.HasPrefix(id.Type, "ISBN") {
			query = append(query, "isbn:"+id.Identifier)
		}
	}

	return
}
