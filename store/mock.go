package store

import (
	"time"
)

var timestamper = time.Now

// UseFrozenTimeStamps sets the time-stamp function to returns a fixed
// time-stamp. It is especially useful for time-sensitive tests and a normal
// user would probably never wants this feature to be set.
func UseFrozenTimeStamps() {
	timestamper = func() time.Time {
		return time.Unix(190701725, 0)
	}
}
