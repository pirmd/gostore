package store

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/blevesearch/bleve/mapping"
)

// Config describes a configuration set for a Store
type Config struct {
	// Path is the path to store the Store
	Path string

	// Logger is the logger used by Store to feedback events.
	// Default to ioutil.Discard (no log).
	Logger *log.Logger

	// TypeField is the name used for identifying record's type to apply
	// specific indexing scheme.
	// Default to "_type"
	TypeField string

	// IndexingAnalyzer is the default analyzer used to index store's records.
	// Available analyzer are any analyzer compatible with bleve-search
	IndexingAnalyzer string

	// IndexingScheme is the bleve's document mapping used to index store's
	// records.
	IndexingScheme *mapping.IndexMappingImpl
}

// NewFromConfig creates a Store from a given Config
func NewFromConfig(cfg *Config) (*Store, error) {
	return New(
		cfg.Path,
		UsingLogger(cfg.Logger),
		UsingDefaultAnalyzer(cfg.IndexingAnalyzer),
		UsingIndexingScheme(cfg.IndexingScheme),
		UsingTypeField(cfg.TypeField),
	)
}

// Options are using a set of variadic functional options for more
// user-friendly api. Idea is coming from
// [[https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis]]

// Option is a function that can tweak the behavior of a Store
type Option func(*Store) error

var logger = log.New(ioutil.Discard, "store ", log.LstdFlags)

// UsingLogger sets the logger used for logging events from store module.
// By default, log messages are discarded (sent to ioutil.Discard)
func UsingLogger(l *log.Logger) Option {
	return func(*Store) error {
		if l != nil {
			logger = l
		}
		return nil
	}
}

// UsingDefaultAnalyzer sets the default anaylzer of the Store's index
//
// Available analyzer are any analyzer compatible with bleve-search.
func UsingDefaultAnalyzer(analyzer string) Option {
	return func(s *Store) error {
		// TODO(pirmd): check what happen if analyzer is empty on bleve's side
		if analyzer != "" {
			s.idx.Mapping.DefaultAnalyzer = analyzer
		}
		return nil
	}
}

// UsingIndexingScheme adds bleve's document mapping to the Store's index
//
// Available mapings are any mappings compatible with bleve-search The index
// mapping can take benefit of Record implementing bleve.Classifier interface.
//
// Mapping only applies to newly created indexes so that you might need to
// manually regenerate the index if you modify the mapping.
func UsingIndexingScheme(idxMappings *mapping.IndexMappingImpl) Option {
	return func(s *Store) error {
		// TODO(pirmd): check what happen on bleve's side if mapping is nil
		if idxMappings != nil {
			s.idx.Mapping = idxMappings
		}
		return nil
	}
}

// UsingTypeField custumizes the name of the field used to identified the type
// of the stored record. Type is used to implement specific indexing scheme
// that can be customized with UsingIndexingScheme.
// Default is "_type".
//
// UsingTypeField shall be used after UsingIndexingScheme
func UsingTypeField(name string) Option {
	return func(s *Store) error {
		if name != "" {
			s.idx.Mapping.TypeField = name
		}
		return nil
	}
}

var timestamper = time.Now

// UsingFrozenTimeStamps sets the time-stamp function to returns a fixed
// time-stamp. It is especially useful for time-sensitive tests and a normal
// user would probably never wants this feature to be set.
func UsingFrozenTimeStamps() Option {
	return func(s *Store) error {
		timestamper = func() time.Time {
			return time.Unix(190701725, 0)
		}
		return nil
	}
}
