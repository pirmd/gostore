package store

import (
	"fmt"
)

type NonBlockingErrors []error

func (c *NonBlockingErrors) Add(err error) {
	*c = append(*c, err)
}

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

func (c *NonBlockingErrors) Err() error {
	if c == nil {
		return nil
	}

	if len(*c) == 0 {
		return nil
	}

	return c
}
