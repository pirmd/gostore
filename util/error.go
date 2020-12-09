package util

import (
	"fmt"
)

// MultiErrors represents a collection of errors.
type MultiErrors []error

// Add adds a new error to the errors collection.
func (c *MultiErrors) Add(err error) {
	if err != nil {
		*c = append(*c, err)
	}
}

// Addf adds a new error to errors collection using fmt.Errorf format.
func (c *MultiErrors) Addf(format string, a ...interface{}) {
	c.Add(fmt.Errorf(format, a...))
}

// Error concatenates all errors in the collection. It allows to satisfy error
// interface.
func (c *MultiErrors) Error() (s string) {
	for i := range *c {
		if i == 0 {
			s = (*c)[i].Error()
		} else {
			s = fmt.Sprintf("%s\n%s", s, (*c)[i])
		}
	}

	return s
}

// Err returns nil if the errors collection is empty or nil, it returns itself
// otherwise.
func (c *MultiErrors) Err() error {
	if c == nil {
		return nil
	}

	if len(*c) == 0 {
		return nil
	}

	return c
}
