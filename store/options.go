package store

// This file gathers the different options available to personnalize any
// Storage
//
// Options are using a set of variadic functional options for more
// user-friendly api Idea is coming from
// [[https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis]]

import (
	"io/ioutil"
	"log"

	"github.com/blevesearch/bleve/mapping"
)

//Option is a function that can tweak the behavior of a Store
type Option func(*Store) error

var logger = log.New(ioutil.Discard, "store ", log.LstdFlags)

//UsingLogger sets the logger used for logging events from store module By
//default, log message are discarded (sent to ioutil.Discard)
func UsingLogger(l *log.Logger) Option {
	return func(*Store) error {
		logger = l
		return nil
	}
}

//UsingDefaultAnalyzer sets the default anaylzer of the Store's index
//
//Available analyzer are any analyzer compatible with bleve-search.
func UsingDefaultAnalyzer(analyzer string) Option {
	return func(s *Store) error {
		s.idx.Mapping.DefaultAnalyzer = analyzer
		return nil
	}
}

//UsingIndexingScheme adds bleve's document mapping to the Store's index
//
//Available mapings are any mappings compatible with bleve-search The index
//mapping can take benefit of Record implementing bleve.Classifier interface.
//
//Mapping only applies to newly created indexes so that you might need to
//manualy regenerate the index if you modify the mapping.
func UsingIndexingScheme(idxMappings *mapping.IndexMappingImpl) Option {
	return func(s *Store) error {
		s.idx.Mapping = idxMappings
		return nil
	}
}

//UsingTypeField custumizes the name of the field used to identified the type
//of the stored record. Default is "_type". Type is used to implement specific
//indexing scheme that can be customized with UsingIndexingScheme
//
//UsingTypeField shall be used after UsingIndexingScheme
func UsingTypeField(name string) Option {
	return func(s *Store) error {
		s.idx.Mapping.TypeField = name
		return nil
	}
}
