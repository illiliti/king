package king

import (
	"context"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/henvic/ctxsignal"
	"github.com/illiliti/king/internal/file"
)

type Preparer interface {
	Prepare(d string) error
}

func (g *Git) Prepare(d string) error {
	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	r, err := git.PlainInit(d, false)

	if err != nil {
		return err
	}

	c := &config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{g.URL},
		Fetch: []config.RefSpec{g.RefSpec},
	}

	if _, err := r.CreateRemote(c); err != nil {
		return err
	}

	if err := r.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: c.Fetch,
		Depth:    1,
		Progress: os.Stderr,
		Tags:     git.AllTags,
	}); err != nil {
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

func (h *HTTP) Prepare(d string) error {
	if h.HasNoExtract {
		return file.CopyFile(h.Path, d)
	}

	return file.ExtractArchive(h.Path, d, 1)
}

func (f *File) Prepare(d string) error {
	if f.IsDir {
		return file.CopyDir(f.Path, d)
	}

	return file.CopyFile(f.Path, d)
}
