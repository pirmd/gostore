package util

import (
	"fmt"

	"github.com/pirmd/text/diff"
)

const (
	majorChangeThreshold = 0.8
)

// HasMajorChanges is true if differences between l and r are deemed as significant.
// Importance of differences is measured field-by-field using a LCS-like distance.
// It is enough for one field to get less than 80% of similar runes to result
// in a true answer.
// Any field from l lacking in r or present in r but lacking in l is not seen
// as a significant change.
func HasMajorChanges(l, r map[string]interface{}) bool {
	if fmt.Sprint(l) == fmt.Sprint(r) {
		return false
	}

	for k := range l {
		if _, exists := r[k]; !exists {
			continue
		}

		valL, valR := fmt.Sprint(l[k]), fmt.Sprint(r[k])
		changes := diff.LCS(valL, valR, diff.ByRunes)

		if s := float32(lenSameDiff(changes)) / float32(maxLen(valL, valR)); s < majorChangeThreshold {
			return true
		}
	}

	return false
}

func lenSameDiff(delta diff.Delta) int {
	if result, isResult := delta.(diff.Result); isResult {
		var same int
		for _, d := range result {
			same += lenSameDiff(d)
		}
		return same
	}

	if delta.Type() == diff.IsSame {
		return len(delta.Value())
	}
	return 0
}

func maxLen(a, b string) int {
	if len(a) > len(b) {
		return len(a)
	}
	return len(b)
}
