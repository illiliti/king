package king

import (
	"context"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/henvic/ctxsignal"
	"github.com/illiliti/king/internal/file"
)

func (g *Git) Prepare(d string) error {
	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	_, err := git.PlainCloneContext(ctx, d, false, &git.CloneOptions{
		URL:          g.URL,
		Depth:        1,
		SingleBranch: true,
		Progress:     os.Stderr,
	})

	return err
}

func (h *HTTP) Prepare(d string) error {
	if h.HasNoExtract {
		return file.CopyFile(h.Path, d)
	}

	return file.Unarchive(h.Path, d, 1)
}

func (f *File) Prepare(d string) error {
	st, err := os.Stat(f.Path)

	if err != nil {
		return err
	}

	if st.IsDir() {
		return file.CopyDir(f.Path, d)
	}

	return file.CopyFile(f.Path, d)
}
