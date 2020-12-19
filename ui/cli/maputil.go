package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pirmd/text/diff"

	"github.com/pirmd/gostore/util"
)

const (
	noOrEmptyValue = "<no value>"
	timeStampFmt   = time.RFC1123Z
	dateFmt        = "2006-01-02"
)

// keyVal represents a collection of (key, values) couples.
type keyVal struct {
	Keys   []string
	Values [][]string
}

func map2kv(maps []map[string]interface{}, fields ...string) *keyVal {
	kv := &keyVal{}

	if len(maps) == 0 {
		return kv
	}

	keys := getKeys(maps, fields...)
	for _, k := range keys {
		if k[0] == '?' {
			kv.Keys = append(kv.Keys, k[1:])
		} else {
			kv.Keys = append(kv.Keys, k)
		}
	}

	for _, m := range maps {
		kv.Values = append(kv.Values, getValues(m, kv.Keys...))
	}

	return kv
}

func (kv *keyVal) KV() [][]string {
	if len(kv.Keys) == 0 {
		return [][]string{}
	}
	return append([][]string{kv.Keys}, kv.Values...)
}

// mergeMaps completes m with n content with the following logic: values of m are
// copied and supersede n values if any. Values of n that are not in m are added.
func mergeMaps(m, n map[string]interface{}) (map[string]interface{}, error) {
	merged := make(map[string]interface{})

	for k, v := range n {
		merged[k] = v
	}

	for k, v := range m {
		merged[k] = v
	}

	return merged, nil
}

// getKeys retrieves key-value couples from a collection of maps. Specific
// notation can be used to identify which key to retrieve:
//  . 'key'   -> field named "key"
//  . '!key'  -> ignore field named "key"
//  . '?keys' -> include field named "key" if value is non null
//  . '*'     -> all remaining fields
//  . '?*'    -> include all remaining fields only if value is non null
func getKeys(maps []map[string]interface{}, fields ...string) (keys []string) {
	if len(fields) == 0 {
		fields = []string{"*"}
	}

	for _, f := range fields {
		switch {
		case f == "*" || f == "?*":
			var allkeys []string
			for _, m := range maps {
				for k := range m {
					if !isInSlice("!"+k, fields) &&
						!isInSlice("?"+k, fields) &&
						!isInSlice(k, fields) &&
						!isInSlice("?"+k, allkeys) &&
						!isInSlice(k, allkeys) {
						if f[0] == '?' {
							if hasValue(k, maps...) {
								allkeys = append(allkeys, "?"+k)
							}
						} else {
							allkeys = append(allkeys, k)
						}
					}
				}
			}
			sort.Strings(allkeys)
			keys = append(keys, allkeys...)

		case f[0] == '!':
			//ignore this field

		case f[0] == '?':
			if hasValue(f[1:], maps...) {
				keys = append(keys, f)
			}

		default:
			keys = append(keys, f)
		}
	}

	return
}

func getValues(m map[string]interface{}, fields ...string) (values []string) {
	for _, f := range fields {
		switch f {
		case "*":
			for k := range m {
				if !isInSlice(k, fields) {
					values = append(values, get(m, k))
				}
			}

		default:
			values = append(values, get(m, f))
		}
	}
	return
}

func get(m map[string]interface{}, key string) string {
	if len(key) == 0 {
		return ""
	}

	k := strings.TrimPrefix(key, "?")
	if v, exists := m[k]; exists {
		if t, ok := v.(time.Time); ok {
			//Only a date
			if strings.HasSuffix(k, "Date") {
				return t.Format(dateFmt)
			}
			//Stamp
			return t.Format(timeStampFmt)
		}

		if !util.IsZero(v) {
			return fmt.Sprintf("%v", v)
		}
	}

	if key[0] != '?' {
		return noOrEmptyValue
	}

	return ""
}

type changeLevel int

const (
	noChange changeLevel = iota
	minorChange
	majorChange

	minorChangeThreshold = 0.8
)

func (lvl *changeLevel) Set(l changeLevel) {
	if (*lvl) < l {
		(*lvl) = l
	}
}

func (lvl changeLevel) String() string {
	return [...]string{"NoChange", "MinorChange", "MajorChange"}[lvl]
}

func hasChanged(l, r map[string]interface{}) (lvl changeLevel) {
	for k := range l {
		if _, exists := r[k]; !exists {
			lvl.Set(minorChange)
			continue
		}

		valL, valR := get(l, k), get(r, k)
		change := diff.Patience(valL, valR, diff.ByWords, diff.ByRunes)

		switch s := float32(sameLen(change)) / float32(maxLen(valL, valR)); {
		case s == 1.0:
			lvl.Set(noChange)
		case s > minorChangeThreshold:
			lvl.Set(minorChange)
		default:
			lvl.Set(majorChange)
		}
	}

	for k := range r {
		if _, exists := l[k]; exists {
			continue
		}

		lvl.Set(majorChange)
	}

	return
}

func hasValue(k string, maps ...map[string]interface{}) bool {
	for _, m := range maps {
		if util.IsZero(m[k]) {
			continue
		}
		return true
	}

	return false
}

func isInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func sameLen(delta diff.Delta) int {
	if result, isResult := delta.(diff.Result); isResult {
		var same int
		for _, d := range result {
			same += sameLen(d)
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
