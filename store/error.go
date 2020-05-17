package store

import (
	"fmt"
)

// NonBlockingErrors represents a collection of errors that don't require to
// stop current activities and can be reported later
type NonBlockingErrors []error

// Add adds a new error to the non-blocking error collection
func (c *NonBlockingErrors) Add(err error) {
	*c = append(*c, err)
}

// Error concatenates all errors in the collection. It allows to satisfy error
// interface
func (c *NonBlockingErrors) Error() (s string) {
	for i := range *c {
		if i == 0 {
			s = (*c)[i].Error()
		} else {
			s = fmt.Sprintf("%s\n%s", s, (*c)[i])
		}
	}

	return s
}

// Err return nil if the non-blocking errors collection is empty or nil, it
// returns itself otherwise
func (c *NonBlockingErrors) Err() error {
	if c == nil {
		return nil
	}

	if len(*c) == 0 {
		return nil
	}

	return c
}
