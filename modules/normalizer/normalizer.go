// Package normalizer is a module that proposes normalized field values based
// on already similar existing records' values.
//
// Limit of use: only support at that time normalization of metadata of string
// (or slice of string) type. I can't think of a situation where you want it to
// happen for other things than strings though.
package normalizer

import (
	"fmt"
	"log"

	"github.com/pirmd/gostore/modules"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "normalizer"
)

var (
	_ modules.Module = (*normalizer)(nil) // Makes sure that we implement modules.Module interface.
)

// Config defines the different module's options.
type Config struct {
	// Fields is the list of fields where you expect normalization to happen.
	// Normalisation is only working for fields featuring strings (or list of
	// strings) like Authors, Publisher...
	Fields []string

	// SimilarityLevel is the measure of similarity that is accepted
	// between two records.
	// .   0: the input text is analyzed first. An attempt is made to use the same
	//        analyzer that was used when the field was indexed.
	// . > 0: the input text is analysed first, the match is done with the given
	//        level of fuzziness.
	// Default to 0.
	SimilarityLevel int
}

func newConfig() *Config {
	return &Config{}
}

type normalizer struct {
	log       *log.Logger
	store     *store.Store
	fields    []string
	fuzziness int
}

func newNormalizer(cfg *Config, logger *log.Logger, store *store.Store) (modules.Module, error) {
	return &normalizer{
		log:       logger,
		store:     store,
		fields:    cfg.Fields,
		fuzziness: cfg.SimilarityLevel,
	}, nil
}

// ProcessRecord looks for pre-existing similar fields value and if any replace
// it in the current Record in hope to normalize its content.
func (n *normalizer) ProcessRecord(r *store.Record) error {
	for _, field := range n.fields {

		switch value := r.Get(field).(type) {
		case nil:

		case []string:
			uniqueValue := make(map[string]struct{})
			for _, v := range value {
				normVal, err := n.normalize(field, v)
				if err != nil {
					return fmt.Errorf("module '%s': fail to normalize: %v", moduleName, err)
				}
				if normVal != nil {
					uniqueValue[normVal.(string)] = struct{}{}
				} else {
					uniqueValue[v] = struct{}{}
				}
			}

			normValue := []string{}
			for k := range uniqueValue {
				normValue = append(normValue, k)
			}
			r.Set(field, normValue)

		case []interface{}:
			uniqueValue := make(map[interface{}]struct{})
			for _, v := range value {
				normVal, err := n.normalize(field, v)
				if err != nil {
					return fmt.Errorf("module '%s': fail to normalize: %v", moduleName, err)
				}
				if normVal != nil {
					uniqueValue[normVal] = struct{}{}
				} else {
					uniqueValue[v] = struct{}{}
				}
			}

			normValue := []interface{}{}
			for k := range uniqueValue {
				normValue = append(normValue, k)
			}
			r.Set(field, normValue)

		case interface{}:
			normVal, err := n.normalize(field, value)
			if err != nil {
				return fmt.Errorf("module '%s': fail to normalize: %v", moduleName, err)
			}
			if normVal != nil {
				r.Set(field, normVal)
			}

		default:
			return fmt.Errorf("module '%s': fail to normalize field %s: incorrect type (%T)", moduleName, field, value)
		}
	}

	return nil
}

// normalize tries to find pre-existing similar field value in the store.
// Return the normed value, empty string otherwise.
func (n *normalizer) normalize(field string, value interface{}) (interface{}, error) {
	n.log.Printf("module '%s': normalizing %s=%+v", moduleName, field, value)

	searchVal, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("incorrect type (%T)", value)
	}

	matches, err := n.store.SearchFields(-1, field, searchVal)
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		n.log.Printf("module '%s': record with the same field value already exist (%+v). Keep initial value.", moduleName, matches)
		return nil, nil
	}

	_, values, err := n.store.MatchFields(n.fuzziness, field, searchVal)
	if err != nil {
		return nil, err
	}

	if len(values[field]) > 0 {
		n.log.Printf("module '%s': find possible similar candidates %+v", moduleName, values[field])
		n.log.Printf("module '%s': select '%v' that has the highest search score", moduleName, values[field][0])
		return values[field][0], nil
	}

	n.log.Printf("module '%s': no similar candidates found. Keep initial value.", moduleName)
	return nil, nil
}

// NewFromRawConfig creates a new module from a raw configuration.
func NewFromRawConfig(rawcfg modules.Unmarshaler, env *modules.Environment) (modules.Module, error) {
	env.Logger.Printf("Module '%s': new module with config '%v'", moduleName, rawcfg)
	cfg := newConfig()

	if err := rawcfg.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("module '%s': bad configuration: %v", moduleName, err)
	}

	return newNormalizer(cfg, env.Logger, env.Store)
}

func init() {
	modules.Register(moduleName, NewFromRawConfig)
}
