package convert

import "fmt"

type PathedError struct {
	Path      []string
	BaseError error
}

func (e *PathedError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.BaseError)
}

func WrapConversionError(prefix string, err error) *PathedError {
	if ce, ok := err.(*PathedError); ok {
		return &PathedError{
			Path:      append([]string{prefix}, ce.Path...),
			BaseError: ce.BaseError,
		}
	}
	return &PathedError{
		Path:      []string{prefix},
		BaseError: err,
	}
}
