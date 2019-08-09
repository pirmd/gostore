package store

import (
	"time"
)

//stampFormats lists all time formats that are recognized by ParseTime
var stampFormats = []string{
	time.RFC3339,
	time.RFC850,
	time.ANSIC,
	"2006",
	"2006-01",
	"2006-01-02",
}

//parseTime parses a time stamp, trying different time format from StampFormats
func parseTime(text string) (t time.Time, err error) {
	for _, fmt := range stampFormats {
		if t, err = time.Parse(fmt, text); err == nil {
			return
		}
	}

	return
}
