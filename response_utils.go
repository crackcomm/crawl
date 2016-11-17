package crawl

import "os"

// WriteResponseFile - Reads response body to memory and writes to file.
func WriteResponseFile(r *Response, fname string) (err error) {
	// Read response body
	body, err := r.Bytes()
	if err != nil {
		return
	}
	f, err := os.Create(fname)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(body)
	if err != nil {
		return
	}
	return
}
