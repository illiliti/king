package king

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/illiliti/king/internal/download"
)

// TODO better docs

type Downloader interface {
	Download(do *DownloadOptions) error
}

type DownloadOptions struct {
	Overwrite bool
	Progress  io.Writer
}

func (h *HTTP) Download(do *DownloadOptions) error {
	if err := do.Validate(); err != nil {
		return fmt.Errorf("validate DownloadOptions: %w", err)
	}

	if !do.Overwrite {
		_, err := os.Stat(h.cs)

		if err == nil {
			return nil
		}

		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	if err := download.Download(h.URL, h.cs, do.Progress); err != nil {
		return fmt.Errorf("download HTTP source %s: %w", h.URL, err)
	}

	return nil
}
