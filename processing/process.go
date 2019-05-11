package processing

import (
	"fmt"

	"github.com/pirmd/gostore/store"
)

var (
	//RecordProcessors contains all known Record's processors
	RecordProcessors = make(map[string]processFn)

	//RecordProcess contains list of Record treatments to be
	//systematically applied to a record before importing it
	//(or updating) into the store
	RecordProcessings = []string{}
)

type processFn func(*store.Record) error

//ProcessRecord applies the specified series of systematic treatments
//to a record. The list of treatments is taken from RecordProcessings
//It fails and exists if a specified processors is unknown or if an
//error occured.
func ProcessRecord(r *store.Record) error {
	for _, p := range RecordProcessings {
		fn, exists := RecordProcessors[p]
		if !exists {
			return fmt.Errorf("process: '%s' is an unknown record processor", p)
		}

		if err := fn(r); err != nil {
			return err
		}
	}

	return nil
}
