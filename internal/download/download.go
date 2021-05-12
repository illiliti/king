package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

type progress struct {
	t string
	n uint64

	s string
	w io.Writer
}

func (p *progress) Write(b []byte) (int, error) {
	n := len(b)
	p.n += uint64(n)

	return n, p.Flush()
}

// TODO print speed, ETA
func (p *progress) Flush() error {
	_, err := fmt.Fprintf(p.w, "\r>> downloading %s [%s%s]", p.s, humanize.Bytes(p.n), p.t)
	return err
}

// TODO remove file if download fails
func Download(s, d string, w io.Writer) error {
	rp, err := http.Get(s)

	if err != nil {
		return err
	}

	defer rp.Body.Close()

	if rp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: non-200 response", s)
	}

	if err := os.MkdirAll(filepath.Dir(d), 0777); err != nil {
		return err
	}

	f, err := os.Create(d)

	if err != nil {
		return err
	}

	p := &progress{
		s: s,
		w: w,
	}

	if rp.ContentLength > 0 {
		p.t = "/" + humanize.Bytes(uint64(rp.ContentLength))
	}

	// TODO print average rate
	defer w.Write([]byte{'\n'})

	if _, err := io.Copy(io.MultiWriter(f, p), rp.Body); err != nil {
		return err
	}

	return f.Close()
}
