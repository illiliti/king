package king

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/henvic/ctxsignal"
)

type Downloader interface {
	Download(force bool) error
}

// TODO progress bar
func (h *HTTP) Download(force bool) error {
	if _, err := os.Stat(h.Path); !force && !os.IsNotExist(err) {
		return nil
	}

	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, h.URL, nil)

	if err != nil {
		return err
	}

	rp, err := http.DefaultClient.Do(rq)

	if err != nil {
		return err
	}

	defer rp.Body.Close()

	if rp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", h.URL, rp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(h.Path), 0777); err != nil {
		return err
	}

	f, err := os.Create(h.Path)

	if err != nil {
		return err
	}

	if _, err := io.Copy(f, rp.Body); err != nil {
		return err
	}

	return f.Close()
}
