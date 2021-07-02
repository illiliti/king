package king

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// TODO better docs

var ErrDownloadAlreadyExist = errors.New("target already downloaded")

type Downloader interface {
	Download(do *DownloadOptions) error
	DownloadContext(ctx context.Context, do *DownloadOptions) error
}

type DownloadOptions struct {
	Overwrite bool
	Progress  io.Writer
}

func (h *HTTP) Download(do *DownloadOptions) error {
	return h.DownloadContext(context.Background(), do)
}

func (h *HTTP) DownloadContext(ctx context.Context, do *DownloadOptions) error {
	if err := do.Validate(); err != nil {
		return fmt.Errorf("validate DownloadOptions: %w", err)
	}

	if err := download(ctx, h.URL, h.cs, do.Overwrite, do.Progress); err != nil {
		return fmt.Errorf("download HTTP source %s: %w", h.URL, err)
	}

	return nil
}

func download(ctx context.Context, s, d string, o bool, p io.Writer) error {
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
		return fmt.Errorf("download %s: %s", s, rp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(d), 0777); err != nil {
		return err
	}

	fl := os.O_WRONLY | os.O_CREATE

	if !o {
		fl |= os.O_EXCL
	}

	f, err := os.OpenFile(d, fl, 0666)

	if err != nil {
		return err
	}

	if _, err := io.Copy(io.MultiWriter(f, p), rp.Body); err != nil {
		if err := os.Remove(d); err != nil {
			return err
		}

		return err
	}

	return f.Close()
}
