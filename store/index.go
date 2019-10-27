package store

import (
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type storeidx struct {
	//Mapping defines specific index mapping
	//Mapping follows bleve's index mapping principles
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

//Open opens or creates a new storeidx
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

//Close cleanly closes the storeidx
func (s *storeidx) Close() error {
	s.idx.Close()
	return nil
}

//Empty removes all content from the index and restart from scratch
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

//Put adds a new value to the new index
func (s *storeidx) Put(r *Record) error {
	return s.idx.Index(r.key, r.value)
}

//Exists checks if an entry  exists for the given key
func (s *storeidx) Exists(key string) (bool, error) {
	doc, err := s.idx.Document(key)
	if err != nil {
		return false, err
	}
	return (doc != nil), nil
}

//Delete suppresses Record from the index
func (s *storeidx) Delete(key string) error {
	return s.idx.Delete(key)
}

//Search looks for known Records'Id regitered in the Index
//The query should follow bleve's query style
func (s *storeidx) Search(query string) (keys []string, err error) {
	q := bleve.NewQueryStringQuery(query)
	results, err := s.idx.Search(bleve.NewSearchRequest(q))
	if err != nil {
		return nil, err
	}

	for _, r := range results.Hits {
		keys = append(keys, r.ID)
	}
	return keys, nil
}

//Walk iterates over all storeidx items and call walkFn for each item
//Walk does not stop if an error is reported by walkFn, such errors will
//be captured and reported back once Walk is over
func (s *storeidx) Walk(walkFn func(string) error) error {
	errWalk := new(NonBlockingErrors)

	//bleve does not support modifying the database during
	//iteration, so we first get all keys, then we act uppon
	//them
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

//TODO: For some reason using NewMatchAllQuery does not return all Documents
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
