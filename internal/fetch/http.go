package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// TODO split into separate packages ?

// TODO progress bar
// TODO signal handling
func HTTPDownload(s, d string) error {
	r, err := http.Get(s)

	if err != nil {
		return err
	}

	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", s, r.Status)
	}

	if err := os.MkdirAll(filepath.Dir(d), 0755); err != nil {
		return err
	}

	f, err := os.Create(d)

	if err != nil {
		return err
	}

	if _, err := io.Copy(f, r.Body); err != nil {
		return err
	}

	return f.Close()
}
