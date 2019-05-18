package books

import (
    "io/ioutil"
    "net/url"
    "net/http"
    "encoding/json"
    "strings"
    "fmt"

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
    Type       string  `json:"type"`
    Identifier string  `json:"identifier"`
}

func (vi *googleVolumeInfo) Isbn13() string {
    for _, id := range vi.Identifier {
        if id.Type == "ISBN_13" {
            return id.Identifier
        }
    }
    return ""
}

func (vi *googleVolumeInfo) PublishedDateAsTime() interface{} {
    if t, err := parseTime(vi.PublishedDate); err == nil {
        return t
    }
    return vi.PublishedDate
}

//Wraps google books api into a Fetcher
type googleBooks struct {}

func (g *googleBooks) LookForBooks(mdata map[string]interface{}) ([]map[string]interface{}, error) {
    queryUrl, err := g.buildQueryUrl(mdata)
    if err != nil {
        return nil, err
    }

	resp, err := http.Get(queryUrl)
	if err != nil {
        return nil, err
	}
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, media.ErrNoMatchFound
    }

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var vol *googleVolumes
    if err := json.Unmarshal(data, &vol); err != nil {
        return nil, err
    }

    metadata := g.vol2mdata(vol)

    return metadata, nil
}

func (g *googleBooks) buildQueryUrl(mdata map[string]interface{}) (string, error) {
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
        return "", fmt.Errorf("Empty query")
    }

    q := url.Values{}
    q.Set("q", strings.Join(query, "+"))
    q.Set("orderBy", "relevance")
    q.Set("printType", "books")
    q.Set("maxResults", "5")     //TODO: config this part
    //TODO: q.Set("langRestrict", "fr") Is it realy needed?

    return googleBooksURL+"?"+q.Encode(), nil
}

func (g *googleBooks) vol2mdata(vol *googleVolumes) (mdata []map[string]interface{}) {
    for _, v := range vol.Items {
        mdata = append(mdata, map[string]interface{}{
            "Title"        : v.VolumeInfo.Title,
            "Authors"      : v.VolumeInfo.Authors,
            "Description"  : v.VolumeInfo.Description,
            "Subject"      : v.VolumeInfo.Subject,
            "ISBN"         : v.VolumeInfo.Isbn13(),
            "Publisher"    : v.VolumeInfo.Publisher,
            "PublishedDate": v.VolumeInfo.PublishedDateAsTime(),
            "PageCount"    : v.VolumeInfo.PageCount,
            "Language"     : v.VolumeInfo.Language,
        })
    }

    return
}
