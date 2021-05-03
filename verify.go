package king

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/hash"

	"github.com/illiliti/king/checksums"
)

// TODO better docs

var (
	ErrSha256NotRegular = errors.New("target must be a regular file")
	ErrVerifyMismatch   = errors.New("checksum mismatch")
)

type Verifier interface {
	Sha256() (string, error) // TODO unexport ?
	Verify() error
}

func (h *HTTP) Sha256() (string, error) {
	x, err := hash.Sha256(h.cs)

	if err == nil {
		return x, nil
	}

	return "", fmt.Errorf("compute sha256 of HTTP source %s: %w", h.cs, err)
}

func (f *File) Sha256() (string, error) {
	st, err := os.Stat(f.Path)

	if err != nil {
		return "", err
	}

	var x string

	if !st.Mode().IsRegular() {
		err = ErrSha256NotRegular
	} else {
		x, err = hash.Sha256(f.Path)
	}

	if err == nil {
		return x, nil
	}

	return "", fmt.Errorf("compute sha256 of File source %s: %w", f.Path, err)
}

func (h *HTTP) Verify() error {
	x, err := h.Sha256()

	if err != nil {
		return err
	}

	if err := verify(h.pkg.Path, x); err != nil {
		return fmt.Errorf("verify HTTP source %s: %w", h.cs, err)
	}

	return nil
}

func (f *File) Verify() error {
	x, err := f.Sha256()

	if errors.Is(err, ErrSha256NotRegular) {
		return nil
	}

	if err != nil {
		return err
	}

	if err := verify(f.pkg.Path, x); err != nil {
		return fmt.Errorf("verify File source %s: %w", f.Path, err)
	}

	return nil
}

func verify(p, x string) error {
	es, err := checksums.Open(filepath.Join(p, "checksums"), os.O_RDONLY)

	if err != nil {
		return err
	}

	defer es.Close()

	if es.Has(x) {
		return nil
	}

	return ErrVerifyMismatch
}
