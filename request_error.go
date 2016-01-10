package crawl

import "fmt"

// RequestError - Crawl error.
type RequestError struct {
	*Request
	Err error
}

// Error - Returns request error message.
func (err *RequestError) Error() string {
	return fmt.Sprintf("%s: %v", err.Request.String(), err.Err)
}
