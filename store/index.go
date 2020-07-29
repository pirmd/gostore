package store

import (
	"fmt"
	"os"

	"github.com/pirmd/gostore/util"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"

	// languages (list from github.com/blevesearch/bleve/blob/master/config/config.go)
	_ "github.com/blevesearch/bleve/analysis/lang/ar"
	_ "github.com/blevesearch/bleve/analysis/lang/bg"
	_ "github.com/blevesearch/bleve/analysis/lang/ca"
	_ "github.com/blevesearch/bleve/analysis/lang/cjk"
	_ "github.com/blevesearch/bleve/analysis/lang/ckb"
	_ "github.com/blevesearch/bleve/analysis/lang/cs"
	_ "github.com/blevesearch/bleve/analysis/lang/da"
	_ "github.com/blevesearch/bleve/analysis/lang/de"
	_ "github.com/blevesearch/bleve/analysis/lang/el"
	_ "github.com/blevesearch/bleve/analysis/lang/en"
	_ "github.com/blevesearch/bleve/analysis/lang/es"
	_ "github.com/blevesearch/bleve/analysis/lang/eu"
	_ "github.com/blevesearch/bleve/analysis/lang/fa"
	_ "github.com/blevesearch/bleve/analysis/lang/fi"
	_ "github.com/blevesearch/bleve/analysis/lang/fr"
	_ "github.com/blevesearch/bleve/analysis/lang/ga"
	_ "github.com/blevesearch/bleve/analysis/lang/gl"
	_ "github.com/blevesearch/bleve/analysis/lang/hi"
	_ "github.com/blevesearch/bleve/analysis/lang/hu"
	_ "github.com/blevesearch/bleve/analysis/lang/hy"
	_ "github.com/blevesearch/bleve/analysis/lang/id"
	_ "github.com/blevesearch/bleve/analysis/lang/in"
	_ "github.com/blevesearch/bleve/analysis/lang/it"
	_ "github.com/blevesearch/bleve/analysis/lang/nl"
	_ "github.com/blevesearch/bleve/analysis/lang/no"
	_ "github.com/blevesearch/bleve/analysis/lang/pt"
	_ "github.com/blevesearch/bleve/analysis/lang/ro"
	_ "github.com/blevesearch/bleve/analysis/lang/ru"
	_ "github.com/blevesearch/bleve/analysis/lang/sv"
	_ "github.com/blevesearch/bleve/analysis/lang/tr"
)

type storeidx struct {
	// Mapping defines specific index mapping.
	// Mapping follows bleve's index mapping principles.
	Mapping *mapping.IndexMappingImpl

	path string
	idx  bleve.Index
}

func newIdx(path string) *storeidx {
	return &storeidx{
		path:    path,
		Mapping: bleve.NewIndexMapping(),
	}
}

// Open opens or creates a new storeidx.
func (s *storeidx) Open() (err error) {
	if s.idx, err = bleve.Open(s.path); err == nil {
		return
	}
	if err != bleve.ErrorIndexPathDoesNotExist {
		return
	}

	s.idx, err = bleve.New(s.path, s.Mapping)
	return
}

// Close cleanly closes the storeidx.
func (s *storeidx) Close() error {
	return s.idx.Close()
}

// Empty removes all content from the index and restart from scratch.
func (s *storeidx) Empty() (err error) {
	if err = s.idx.Close(); err != nil {
		return
	}

	if err = os.RemoveAll(s.path); err != nil {
		return
	}

	err = s.Open()
	return
}

// Put adds a new value to the new index.
func (s *storeidx) Put(r *Record) error {
	return s.idx.Index(r.key, r.Value())
}

// Get retrieves a value from the index.
func (s *storeidx) Get(key string) (*Record, error) {
	// Largely inspired by github.com/blevesearch/bleve/http/doc_get.go
	data := make(map[string]interface{})

	doc, err := s.idx.Document(key)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, fmt.Errorf("no such document '%s'", key)
	}

	for _, field := range doc.Fields {
		var val interface{}

		switch field := field.(type) {
		case *document.TextField:
			val = string(field.Value())

		case *document.NumericField:
			if val, err = field.Number(); err != nil {
				return nil, fmt.Errorf("failed to retrieve numeric value for field '%s'[%s]: %s", key, field.Name(), err)
			}

		case *document.DateTimeField:
			if val, err = field.DateTime(); err != nil {
				return nil, fmt.Errorf("failed to retrieve date value for field '%s'[%s]: %s", key, field.Name(), err)
			}
		}

		existing, existed := data[field.Name()]
		if existed {
			switch existing := existing.(type) {
			case []interface{}:
				data[field.Name()] = append(existing, val)
			case interface{}:
				data[field.Name()] = []interface{}{existing, val}
			}
		} else {
			data[field.Name()] = val
		}
	}

	return NewRecord(key, data), nil
}

// Exists checks if an entry  exists for the given key.
func (s *storeidx) Exists(key string) (bool, error) {
	doc, err := s.idx.Document(key)
	if err != nil {
		return false, err
	}
	return (doc != nil), nil
}

// Delete suppresses Record from the index.
func (s *storeidx) Delete(key string) error {
	return s.idx.Delete(key)
}

// Search looks for Records' keys registered in the Index that match the query.
// The query follows bleve's query syntax (http://blevesearch.com/docs/Query-String-Query/)
func (s *storeidx) Search(query string) ([]string, error) {
	q := bleve.NewQueryStringQuery(query)
	searchRequest := bleve.NewSearchRequest(q)

	results, err := s.idx.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, r := range results.Hits {
		keys = append(keys, r.ID)
	}
	return keys, nil
}

// SearchFields looks for Records' keys registered in the Index that match the
// provided fields value.  Level of accepted fuzziness can be specified.
func (s *storeidx) SearchFields(fields map[string]interface{}, fuzziness int) ([]string, error) {
	var queries []query.Query
	for field, match := range fields {
		if match != nil {
			q := bleve.NewMatchQuery(match.(string))
			q.SetField(field)
			q.SetFuzziness(fuzziness)
			queries = append(queries, q)
		}
	}

	q := bleve.NewConjunctionQuery(queries...)

	searchRequest := bleve.NewSearchRequest(q)

	results, err := s.idx.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, r := range results.Hits {
		keys = append(keys, r.ID)
	}
	return keys, nil
}

// Walk iterates over all storeidx items and call walkFn for each item.
// Walk does not stop if an error is reported by walkFn, such errors will
// be captured and reported back once Walk is over.
func (s *storeidx) Walk(walkFn func(string) error) error {
	errWalk := new(util.MultiErrors)

	//bleve does not support modifying the database during
	//iteration, so we first get all keys, then we act upon
	//them.
	keys, err := s.matchAll()
	if err != nil {
		return err
	}

	for _, k := range keys {
		if err := walkFn(k); err != nil {
			errWalk.Add(err)
		}
	}

	return errWalk.Err()
}

// Fields lists the indexed fields.
func (s *storeidx) Fields() ([]string, error) {
	return s.idx.Fields()
}

// matchAll retrieves all known records.
// TODO: For some reason using NewMatchAllQuery does not return all Documents
//      So that I use this func that seems to work. No time to investigate
func (s *storeidx) matchAll() ([]string, error) {
	idx, _, err := s.idx.Advanced()
	if err != nil {
		return nil, err
	}
	idxReader, err := idx.Reader()
	if err != nil {
		return nil, err
	}
	defer idxReader.Close()

	reader, err := idxReader.DocIDReaderAll()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	keys := []string{}
	id, err := reader.Next()
	if err != nil {
		return nil, err
	}

	for id != nil {
		i, err := idxReader.ExternalID(id)
		if err != nil {
			return nil, err
		}

		keys = append(keys, i)

		id, err = reader.Next()
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}
