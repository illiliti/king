package king

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/manifest"
)

// TODO String() ?
type Alternative struct {
	Name string
	Path string

	cfg *Config
}

func (c *Config) NewAlternativeByPath(p string) (*Alternative, error) {
	dd, err := os.ReadDir(filepath.Join(c.RootDir, ChoicesDir))

	if err != nil {
		return nil, err
	}

	s := strings.ReplaceAll(p, "/", ">")

	for _, de := range dd {
		if !strings.HasSuffix(de.Name(), s) {
			continue
		}

		return &Alternative{
			Name: strings.TrimSuffix(de.Name(), s),
			Path: p,
			cfg:  c,
		}, nil
	}

	return nil, fmt.Errorf("alternative %s: not found", p)
}

func (c *Config) NewAlternativeByNamePath(n, p string) (*Alternative, error) {
	a := filepath.Join(c.RootDir, ChoicesDir, n+strings.ReplaceAll(p, "/", ">"))
	st, err := os.Lstat(a)

	if err != nil {
		return nil, err
	}

	if st.IsDir() {
		return nil, fmt.Errorf("alternative %s: is a directory", a)
	}

	return &Alternative{
		Name: n,
		Path: p,
		cfg:  c,
	}, nil
}

func (a *Alternative) Swap() (*Alternative, error) {
	sp, err := a.cfg.NewPackageByName(Sys, a.Name)

	if err != nil {
		return nil, err
	}

	cp, err := a.cfg.NewPackageByPath(a.Path)

	if err != nil {
		return nil, err
	}

	defer pathsOnce.Reset()

	ap := strings.ReplaceAll(a.Path, "/", ">")

	if err := replace(cp, true, a.Path, filepath.Join(ChoicesDir, cp.Name+ap)); err != nil {
		return nil, err
	}

	if err := replace(sp, false, filepath.Join(ChoicesDir, sp.Name+ap), a.Path); err != nil {
		return nil, err
	}

	return &Alternative{
		Name: cp.Name,
		Path: a.Path,
		cfg:  a.cfg,
	}, nil
}

func replace(p *Package, c bool, f, t string) error {
	if err := os.Rename(filepath.Join(p.cfg.RootDir, f), filepath.Join(p.cfg.RootDir, t)); err != nil {
		return err
	}

	m, err := manifest.Open(filepath.Join(p.Path, "manifest"))

	if err != nil {
		return err
	}

	m.Replace(f, t)

	if c {
		m.Insert(ChoicesDir + "/")
	} else {
		m.Delete(ChoicesDir + "/")
	}

	if err := m.Flush(); err != nil {
		return err
	}

	return m.Close()
}
