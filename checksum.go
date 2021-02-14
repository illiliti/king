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
	ErrIsDir = errors.New("is a directory")
)

type Checksum interface {
	Sha256() (string, error)
	Verify() error
}

func (h *HTTP) Sha256() (string, error) {
	return file.Sha256(h.Path)
}

func (f *File) Sha256() (string, error) {
	st, err := os.Stat(f.Path)

	if err != nil {
		return "", err
	}

	if st.IsDir() {
		return "", fmt.Errorf("checksum %s: %w", f.Path, ErrIsDir)
	}

	return file.Sha256(f.Path)
}

func (h *HTTP) Verify() error {
	x, err := h.Sha256()

	if err != nil {
		return err
	}

	return verify(h.pkg, h.Path, x)
}

func (f *File) Verify() error {
	x, err := f.Sha256()

	if err != nil {
		if errors.Is(err, ErrIsDir) {
			return nil
		}

		return err
	}

	return verify(f.pkg, f.Path, x)
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
