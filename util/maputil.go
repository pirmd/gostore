package util

import (
	"fmt"
	"sort"
)

// Sort sorts collection of map[string]interface{} according to the provided
// sort order.
func Sort(maps []map[string]interface{}, sortBy []string) []map[string]interface{} {
	sorted := make([]map[string]interface{}, len(maps))
	copy(sorted, maps)

	sort.Slice(sorted, func(i, j int) bool {
		for _, f := range sortBy {
			vi, vj := fmt.Sprintf("%v", sorted[i][f]), fmt.Sprintf("%v", sorted[j][f])

			if vi == vj {
				continue
			}
			return (vi < vj)
		}

		return false
	})

	return sorted
}

// CopyMap creates a deep copy of map.
func CopyMap(m map[string]interface{}) map[string]interface{} {
	mcopied := map[string]interface{}{}

	for k, v := range m {
		mapval, isMap := v.(map[string]interface{})
		if isMap {
			mcopied[k] = CopyMap(mapval)
			continue
		}

		sliceval, isSlice := v.([]interface{})
		if isSlice {
			mcopied[k] = CopySlice(sliceval)
			continue
		}

		mcopied[k] = v
	}

	return mcopied
}

// CopySlice creates a deep copy of a slice.
func CopySlice(s []interface{}) []interface{} {
	scopied := []interface{}{}

	for _, v := range s {
		mapval, isMap := v.(map[string]interface{})
		if isMap {
			scopied = append(scopied, CopyMap(mapval))
			continue
		}

		sliceval, isSlice := v.([]interface{})
		if isSlice {
			scopied = append(scopied, CopySlice(sliceval))
			continue
		}

		scopied = append(scopied, v)
	}

	return scopied
}
