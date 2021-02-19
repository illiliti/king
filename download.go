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

// Downloader abstracts downloadable sources.
type Downloader interface {
	Download(force bool) error
}

// Download downloads http source into SourceDir + package
// name + ExtractDir (if non-empty) + basename of URL source
//
// TODO progress bar
func (h *HTTP) Download(force bool) error {
	if !force {
		_, err := os.Stat(h.p)

		if err == nil {
			return nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, h.u, nil)

	if err != nil {
		return err
	}

	rp, err := http.DefaultClient.Do(rq)

	if err != nil {
		return err
	}

	defer rp.Body.Close()

	if rp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", h.u, rp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(h.p), 0777); err != nil {
		return err
	}

	f, err := os.Create(h.p)

	if err != nil {
		return err
	}

	if _, err := io.Copy(f, rp.Body); err != nil {
		return err
	}

	return f.Close()
}
