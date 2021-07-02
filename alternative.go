package king

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/manifest"
)

// TODO unit tests
// TODO better docs

var ErrAlternativeNotFound = errors.New("conflict has not been occured yet")

// Alternative represents conflict between multiple packages.
//
// See https://k1sslinux.org/package-manager#3.2
type Alternative struct {
	Name string
	Path string

	cfg *Config
}

type AlternativeOptions struct {
	Name string
	Path string
}

func NewAlternative(c *Config, ao *AlternativeOptions) (*Alternative, error) {
	if err := ao.Validate(); err != nil {
		return nil, fmt.Errorf("validate AlternativeOptions: %w", err)
	}

	aa, err := Alternatives(c)

	if err != nil {
		return nil, err
	}

	for _, a := range aa {
		if a.Path != ao.Path {
			continue
		}

		if ao.Name != "" && a.Name != ao.Name {
			continue
		}

		return a, nil
	}

	return nil, fmt.Errorf("find alternative %s: %w", ao.Path, ErrAlternativeNotFound)
}

func Alternatives(c *Config) ([]*Alternative, error) {
	// TODO cache results

	dd, err := os.ReadDir(c.AlternativeDir)

	if err != nil {
		return nil, err
	}

	aa := make([]*Alternative, 0, len(dd))

	for _, de := range dd {
		// TODO panic()
		if de.IsDir() {
			continue
		}

		i := strings.Index(de.Name(), ">")

		// TODO panic()
		if i < 1 {
			continue
		}

		aa = append(aa, &Alternative{
			Name: de.Name()[:i],
			Path: strings.ReplaceAll(de.Name()[i:], ">", "/"),
			cfg:  c,
		})
	}

	return aa, nil
}

func (a *Alternative) Swap() (*Alternative, error) {
	sp, err := NewPackage(a.cfg, &PackageOptions{
		Name: a.Name,
		From: Database,
	})

	if err != nil {
		return nil, err
	}

	cp, err := NewPackage(a.cfg, &PackageOptions{
		Path: a.Path,
	})

	if err != nil {
		return nil, err
	}

	defer a.cfg.ResetOwnedPaths()

	ap := strings.ReplaceAll(a.Path, "/", ">")

	if err := swap(cp, a.Path, filepath.Join(a.cfg.ad, cp.Name+ap), true); err != nil {
		return nil, fmt.Errorf("swap path %s: %w", a.Path, err)
	}

	if err := swap(sp, filepath.Join(a.cfg.ad, sp.Name+ap), a.Path, false); err != nil {
		return nil, fmt.Errorf("swap path %s: %w", a.Path, err)
	}

	return &Alternative{
		Name: cp.Name,
		Path: a.Path,
		cfg:  a.cfg,
	}, nil
}

func swap(p *Package, s, d string, ad bool) error {
	if err := os.Rename(filepath.Join(p.cfg.RootDir, s), filepath.Join(p.cfg.RootDir, d)); err != nil {
		return err
	}

	mf, err := manifest.Open(filepath.Join(p.Path, "manifest"), os.O_RDWR)

	if err != nil {
		return err
	}

	mf.Replace(s, d)

	if ad {
		mf.Insert(p.cfg.ad + "/")
	} else {
		mf.Delete(p.cfg.ad + "/")
	}

	if err := mf.Flush(); err != nil {
		return err
	}

	return mf.Close()
}

func (a *Alternative) String() string {
	return a.Name + " " + a.Path
}
