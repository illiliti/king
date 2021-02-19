package king

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/chksum"
	"github.com/illiliti/king/internal/file"
)

var (
	// ErrIsDir indicates that file is a directory.
	ErrIsDir = errors.New("is a directory")
)

// Checksum abstracts checksum-related functions for the source.
type Checksum interface {
	Sha256() (string, error)
	Verify() error
}

// Sha256 returns sha256 sum of http source
func (h *HTTP) Sha256() (string, error) {
	return file.Sha256(h.p)
}

// Sha256 returns sha256 sum of file source
func (f *File) Sha256() (string, error) {
	st, err := os.Stat(f.p)

	if err != nil {
		return "", err
	}

	if st.IsDir() {
		return "", fmt.Errorf("checksum %s: %w", f.p, ErrIsDir)
	}

	return file.Sha256(f.p)
}

// Verify checks if sha256 sum of http source contains in checksums file
func (h *HTTP) Verify() error {
	x, err := h.Sha256()

	if err != nil {
		return err
	}

	return verify(h.pkg, h.p, x)
}

// Verify checks if sha256 sum of file source contains in checksums file
func (f *File) Verify() error {
	x, err := f.Sha256()

	if errors.Is(err, ErrIsDir) {
		return nil
	}

	if err != nil {
		return err
	}

	return verify(f.pkg, f.p, x)
}

func verify(p *Package, s, x string) error {
	c, err := chksum.Open(filepath.Join(p.Path, "checksums"))

	if err != nil {
		return err
	}

	if c.HasEntry(x) {
		return nil
	}

	return fmt.Errorf("verify %s: mismatch %s", s, x)
}
