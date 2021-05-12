package king

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/hash"

	"github.com/illiliti/king/etcsums"
	"github.com/illiliti/king/manifest"
)

// TODO unit tests
// TODO better docs

var (
	ErrRemoveUnresolvedDependencies = errors.New("other packages depend on target")
)

// RemoveOptions provides facilities for removing package.
type RemoveOptions struct {
	// NoCheckReverseDependencies forcefully removes package even if other
	// installed packages depends on it.
	NoCheckReverseDependencies bool

	// RemoveEtcFiles forcefully removes /etc/* files without special handling
	RemoveEtcFiles bool
}

// Remove purges package from the system and gracefully swaps dangling
// alternatives if they are exist.
func (p *Package) Remove(ro *RemoveOptions) error {
	if !ro.NoCheckReverseDependencies {
		if err := unresolvedDependencies(p); err != nil {
			return err
		}
	}

	mf, es, err := openManifestEtcsums(p.Path, ro.RemoveEtcFiles, false)

	if err != nil {
		return err
	}

	if err := swapAlternatives(p, mf); err != nil {
		return fmt.Errorf("swap alternatives: %w", err)
	}

	defer p.cfg.ResetOwnedPaths()
	defer p.cfg.ResetReverseDependencies()

	for _, r := range mf.Sort(manifest.Files) {
		if err := remove(es, p.cfg.RootDir, r); err != nil {
			return err
		}
	}

	return nil
}

func unresolvedDependencies(p *Package) error {
	dd, err := p.ReverseDependencies()

	if errors.Is(err, ErrReverseDependenciesNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	return fmt.Errorf("remove package %s: %w: %s", p.Name, ErrRemoveUnresolvedDependencies, dd)
}

func swapAlternatives(p *Package, mf *manifest.Manifest) error {
	for _, r := range mf.Sort(manifest.NoSort) {
		a, err := NewAlternative(p.cfg, &AlternativeOptions{
			Path: r,
		})

		if err != nil {
			continue
		}

		if a.Name == p.Name {
			continue
		}

		if _, err := a.Swap(); err != nil {
			return err
		}
	}

	if err := mf.Rehash(); err != nil {
		return err
	}

	return mf.Close()
}

func remove(es *etcsums.Etcsums, r, p string) error {
	rp := filepath.Join(r, p)
	st, err := os.Lstat(rp)

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	isEmpty := func() bool {
		f, _ := os.Open(rp) // TODO panic on error
		defer f.Close()

		_, err := f.ReadDir(1)
		return errors.Is(err, io.EOF)
	}

	switch {
	case st.IsDir() && !isEmpty():
		return nil
	case es != nil && st.Mode().IsRegular() && strings.HasPrefix(p, "/etc/"):
		x, err := hash.Sha256(rp)

		if err != nil {
			return err
		}

		if !es.Has(x) {
			return nil
		}
	}

	return os.Remove(rp)
}
