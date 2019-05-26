package books

import (
	"time"
)

//stampFormats lists all time formats that are recognized by ParseTime
var stampFormats = []string{
	"1976",
	"1976-01",
	"1976-01-17",
	time.RFC850,
	time.ANSIC,
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
