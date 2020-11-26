package king

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/file"
)

var (
	ErrIsDir = errors.New("is a directory")
)

type Checksumer interface {
	Checksum() (string, error)
	Verify() error

	// TODO
	// Flush() error
	// Save() error
}

func (h *HTTP) Checksum() (string, error) {
	return file.Sha256Sum(h.Path)
}

func (f *File) Checksum() (string, error) {
	if f.IsDir {
		return "", fmt.Errorf("checksum %s: %w", f.Path, ErrIsDir)
	}

	return file.Sha256Sum(f.Path)
}

func (h *HTTP) Verify() error {
	x, err := h.Checksum()

	if err != nil {
		return err
	}

	return verify(h.pkg, x)
}

func (f *File) Verify() error {
	x, err := f.Checksum()

	if err != nil {
		if errors.Is(err, ErrIsDir) {
			return nil
		}

		return err
	}

	return verify(f.pkg, x)
}

func verify(p *Package, x string) error {
	err := p.checksumsOnce.Do(func() error {
		f, err := os.Open(filepath.Join(p.Path, "checksums"))

		if err != nil {
			return err
		}

		p.checksums = make(map[string]bool)
		sc := bufio.NewScanner(f)

		for sc.Scan() {
			p.checksums[strings.Fields(sc.Text())[0]] = true
		}

		return sc.Err()
	})

	if err != nil {
		return err
	}

	if p.checksums[x] {
		return nil
	}

	return fmt.Errorf("checksum mismatch: %s", x)
}
