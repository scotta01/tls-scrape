package scraper

import "fmt"

// MultiError is a custom error type that encapsulates multiple errors
// with their associated domain.
type MultiError struct {
	Errors map[string]error
}

// Error returns a string representation of the MultiError. It aggregates
// the individual error messages for each domain.
func (me *MultiError) Error() string {
	errMsg := "Multiple errors occurred:\n"
	for domain, err := range me.Errors {
		errMsg += fmt.Sprintf("Domain: %s, Error: %s\n", domain, err.Error())
	}
	return errMsg
}
