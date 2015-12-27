package open

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/crackcomm/crawl"
	"github.com/satori/go.uuid"
	"github.com/skratchdot/open-golang/open"
)

// Open - Opens crawl response in browser.
func Open(resp *crawl.Response) error {
	fname := filepath.Join(os.TempDir(), fmt.Sprintf("%s.html", uuid.NewV4().String()))
	body, err := resp.ReadBody()
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(fname, body, os.ModePerm); err != nil {
		return err
	}
	return open.Start(fmt.Sprintf("file://%s", fname))
}
