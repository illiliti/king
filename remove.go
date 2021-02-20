package king

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/illiliti/king/internal/etcsum"
	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/manifest"
)

// Remove removes given package from the system.
//
// If force equal false, Remove checks if other packages depends
// on given package and returns error if so. Otherwise, this check is
// entirely skipped.
func (p *Package) Remove(force bool) error {
	if !force {
		dd, err := p.ReverseDepends()

		if err == nil {
			return fmt.Errorf("remove %s: required by %s", p.Name, dd)
		}
	}

	om, err := manifest.Open(filepath.Join(p.Path, "manifest"))

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	defer om.Close()

	oe, err := etcsum.Open(filepath.Join(p.Path, "etcsums"))

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	defer oe.Close()

	for _, r := range om.Remove() {
		oa, err := NewAlternativeByPath(p.cfg, r)

		if err != nil {
			continue
		}

		if oa.Name == p.Name {
			continue
		}

		if _, err := oa.Swap(); err != nil {
			return err
		}
	}

	if err := om.Rehash(); err != nil {
		return err
	}

	defer atomic.StoreUint32(&pathsCount, 0)
	defer atomic.StoreUint32(&dependenciesCount, 0)

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	for _, r := range om.Remove() {
		rp := filepath.Join(p.cfg.RootDir, r)
		st, err := os.Lstat(rp)

		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return err
		}

		switch {
		case st.Mode().IsRegular() && strings.HasPrefix(r, "/etc/"):
			x, err := file.Sha256(rp)

			if err != nil {
				return err
			}

			if !oe.HasEntry(x) {
				continue
			}
		case st.IsDir():
			err := func() error {
				f, err := os.Open(rp)

				if err != nil {
					return err
				}

				defer f.Close()

				_, err = f.ReadDir(1)
				return err
			}()

			if !errors.Is(err, io.EOF) {
				if err == nil {
					continue
				}

				return err
			}
		}

		if err := os.Remove(rp); err != nil {
			return err
		}
	}

	return nil
}
