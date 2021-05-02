package king

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/illiliti/king/internal/archive"
	"github.com/illiliti/king/internal/cp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

// TODO better docs

// Extractor represents interface of extractable source.
type Extractor interface {
	// Extract(d string) (es *ExtractState, error)

	Extract(d string) error
	ExtractDir() string
}

// type ExtractOptions struct {
// 	Destination string
// 	Progress    io.Writer
// }

// Extract clones git source into specified directory.
func (g *Git) Extract(d string) error {
	if err := gitFetch(g.URL, d, g.rs); err != nil {
		return fmt.Errorf("fetch Git source %s: %w", g.URL, err)
	}

	return nil
}

func gitFetch(s, d string, rs []config.RefSpec) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	r, err := git.PlainInit(d, false)

	if err != nil {
		return err
	}

	u, err := r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{s},
	})

	if err != nil {
		return err
	}

	err = u.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: rs,
		Depth:    1,
		Tags:     git.AllTags,
	})

	if err != nil {
		return err
	}

	h, err := r.Head()

	if err != nil {
		return err
	}

	w, err := r.Worktree()

	if err != nil {
		return err
	}

	return w.Checkout(&git.CheckoutOptions{
		Hash: h.Hash(),
	})
}

// Extract copies file to specified directory.
func (f *File) Extract(d string) error {
	st, err := os.Stat(f.Path)

	if err != nil {
		return err
	}

	if err := os.MkdirAll(d, 0777); err != nil {
		return err
	}

	// TODO symlinks
	if st.IsDir() {
		err = cp.CopyDir(f.Path, d)
	} else {
		err = cp.CopyFile(f.Path, d)
	}

	if err == nil {
		return nil
	}

	return fmt.Errorf("copy File source %s: %w", f.Path, err)
}

// Extract copies (or extracts - it depends on ?no-extract flag) source
// into specified directory.
func (h *HTTP) Extract(d string) error {
	if h.ne {
		if err := os.MkdirAll(d, 0777); err != nil {
			return err
		}

		if err := cp.CopyFile(h.cs, d); err != nil {
			return fmt.Errorf("copy HTTP source %s: %w", h.cs, err)
		}
	}

	if err := archive.Extract(h.cs, d, 1); err != nil {
		return fmt.Errorf("extract HTTP source %s: %w", h.cs, err)
	}

	return nil
}

// ExtractDir returns extraction directory (second field within sources file)
func (g *Git) ExtractDir() string {
	return g.ed
}

// ExtractDir returns extraction directory (second field within sources file)
func (h *HTTP) ExtractDir() string {
	return h.ed
}

// ExtractDir returns extraction directory (second field within sources file)
func (f *File) ExtractDir() string {
	return f.ed
}
