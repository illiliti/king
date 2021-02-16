package king

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
)

type Downloader interface {
	Download(force bool) error
}

// TODO progress bar
func (h *HTTP) Download(force bool) error {
	if _, err := os.Stat(h.Path); !force && !errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

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
		return fmt.Errorf("download %s: %s", h.URL, rp.Status)
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
