package king

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-git/go-git/v5"
	"github.com/illiliti/king/internal/file"
)

// Extract clones git source into specified directory.
func (g *Git) Extract(d string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	_, err := git.PlainCloneContext(ctx, d, false, &git.CloneOptions{
		URL:          g.u,
		Depth:        1,
		SingleBranch: true,
		Progress:     os.Stderr,
	})

	return err
}

// Extract copies/extracts file source into specified directory.
func (f *File) Extract(d string) error {
	st, err := os.Stat(f.p)

	if err != nil {
		return err
	}

	if st.IsDir() {
		return file.CopyDir(f.p, d)
	}

	if f.ia {
		return file.Unarchive(f.p, d, 1)
	}

	return file.CopyFile(f.p, d)
}

// Extract copies/extracts http source into specified directory.
func (h *HTTP) Extract(d string) error {
	if h.ne {
		return file.CopyFile(h.p, d)
	}

	return file.Unarchive(h.p, d, 1)
}

// ExtractDir returns additional (relative to the build directory) directory
func (g *Git) ExtractDir() string {
	return g.d
}

// ExtractDir returns additional (relative to the build directory) directory
func (h *HTTP) ExtractDir() string {
	return h.d
}

// ExtractDir returns additional (relative to the build directory) directory
func (f *File) ExtractDir() string {
	return f.d
}
