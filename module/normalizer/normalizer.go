package normalizer

import (
	"fmt"

	"github.com/pirmd/gostore/module"
	"github.com/pirmd/gostore/store"
)

const (
	moduleName = "normalizer"
)

var (
	_ module.Module = (*normalizer)(nil) // Makes sure that we implement module.Module interface.
)

type config struct {
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

func newConfig() module.Factory {
	return &config{}
}

func (cfg *config) NewModule(env *module.Environment) (module.Module, error) {
	return newNormalizer(cfg, env)
}

// normalizer is a gostore's module that proposes normalized field values based
// on already similar existing records' values.
//
// Limit of use: only support at that time normalization of metadata of string
// (or slice of string) type. I can't think of a situation where you want it to
// happen for other things than strings though.
type normalizer struct {
	*module.Environment
	fields    []string
	fuzziness int
}

func newNormalizer(cfg *config, env *module.Environment) (*normalizer, error) {
	return &normalizer{
		Environment: env,
		fields:      cfg.Fields,
		fuzziness:   cfg.SimilarityLevel,
	}, nil
}

// Process looks for pre-existing similar fields value and if any replace
// it in the current Record in hope to normalize its content.
func (n *normalizer) Process(r *store.Record) error {
	for _, field := range n.fields {
		switch value := r.Get(field).(type) {
		case nil:

		case []string:
			var normValues []string

			uniqueValues := make(map[interface{}]struct{})
			for _, v := range value {
				normVal, err := n.normalize(field, v)
				if err != nil {
					return fmt.Errorf("module '%s': fail to normalize: %v", moduleName, err)
				}

				if _, exist := uniqueValues[normVal]; exist {
					continue
				}

				normValues = append(normValues, normVal.(string))
				uniqueValues[normVal] = struct{}{}
			}

			r.Set(field, normValues)

		case []interface{}:
			var normValues []interface{}

			uniqueValues := make(map[interface{}]struct{})
			for _, v := range value {
				normVal, err := n.normalize(field, v)
				if err != nil {
					return fmt.Errorf("module '%s': fail to normalize: %v", moduleName, err)
				}

				if _, exist := uniqueValues[normVal]; exist {
					continue
				}

				normValues = append(normValues, normVal)
				uniqueValues[normVal] = struct{}{}
			}
			r.Set(field, normValues)

		case interface{}:
			normVal, err := n.normalize(field, value)
			if err != nil {
				return fmt.Errorf("module '%s': fail to normalize: %v", moduleName, err)
			}
			r.Set(field, normVal)

		default:
			return fmt.Errorf("module '%s': fail to normalize field %s: incorrect type (%T)", moduleName, field, value)
		}
	}

	return nil
}

// normalize tries to find pre-existing similar field value in the store.
// Return the original value if nothing is found.
func (n *normalizer) normalize(field string, value interface{}) (interface{}, error) {
	n.Logger.Printf("module '%s': normalizing %s=%+v", moduleName, field, value)

	searchVal, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("incorrect type (%T)", value)
	}

	matches, err := n.Store.SearchFields(-1, field, searchVal)
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		n.Logger.Printf("module '%s': record with the same field value already exist (%+v). Keep initial value.", moduleName, matches)
		return nil, nil
	}

	_, values, err := n.Store.MatchFields(n.fuzziness, field, searchVal)
	if err != nil {
		return nil, err
	}

	if len(values[field]) > 0 {
		n.Logger.Printf("module '%s': find possible similar candidates %+v", moduleName, values[field])
		n.Logger.Printf("module '%s': select '%v' that has the highest search score", moduleName, values[field][0])
		return values[field][0], nil
	}

	n.Logger.Printf("module '%s': no similar candidates found. Keep initial value.", moduleName)
	return value, nil
}

func init() {
	module.Register(moduleName, newConfig)
}
