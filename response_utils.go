package crawl

import (
	"io"
	"os"
)

// WriteResponseFile - Reads response body to memory and writes to file.
func WriteResponseFile(r *Response, fname string) (err error) {
	// Read response body
	body, err := r.BodyBuffer()
	if err != nil {
		return
	}
	f, err := os.Create(fname)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = io.Copy(f, body)
	if err != nil {
		return
	}
	return
}
