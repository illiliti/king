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

	"github.com/illiliti/king/internal/etcsum"
	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/manifest"
)

func (p *Package) Remove(force bool) error {
	dd, err := p.ReverseDepends()

	if !force && err == nil && len(dd) > 0 {
		return fmt.Errorf("remove %s: required by %s", p.Name, strings.Join(dd, ", "))
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
	defer pathsOnce.Reset()
	defer dependenciesOnce.Reset()

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	for _, r := range om.Remove() {
		oa, err := p.cfg.NewAlternativeByPath(r)

		if err != nil {
			continue
		}

		if oa.Name == p.Name {
			continue
		}

		na, err := oa.Swap()

		if err != nil {
			return err
		}

		om.Insert(ChoicesDir)
		om.Replace(na.Path, filepath.Join(ChoicesDir, na.Name+strings.ReplaceAll(na.Path, "/", ">")))
	}

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
			f, err := os.Open(rp)

			if err != nil {
				return err
			}

			_, err = f.ReadDir(1)

			if err := f.Close(); err != nil {
				return err
			}

			if !errors.Is(err, io.EOF) {
				continue
			}
		}

		if err := os.Remove(rp); err != nil {
			return err
		}
	}

	return nil
}
