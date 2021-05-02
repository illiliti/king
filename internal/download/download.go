package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
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

func Download(s, d string, w io.Writer) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, s, nil)

	if err != nil {
		return err
	}

	rp, err := http.DefaultClient.Do(rq)

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
