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
