package king

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/once"
)

// PackageType represepents package type
type PackageType uint

const (
	// Any is the same as Sys + Usr
	Any PackageType = iota

	// Sys finds package in SysDB
	Sys

	// Usr finds package in UserDB
	Usr
)

var (
	paths     map[string]*Package
	pathsOnce once.Once
)

// Package represents location to package.
//
// See https://kiss.armaanb.net/package-system#1.0
type Package struct {
	Name string
	Path string

	cfg *Config
}

// NewPackageByName returns a pointer to Package with appropriate type.
func NewPackageByName(c *Config, t PackageType, n string) (*Package, error) {
	newPackage := func(n string, dd ...string) (*Package, error) {
		for _, db := range dd {
			p := filepath.Join(db, n)
			st, err := os.Stat(p)

			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			if err != nil {
				return nil, err
			}

			if !st.IsDir() {
				continue
			}

			return &Package{
				Name: n,
				Path: p,
				cfg:  c,
			}, nil
		}

		return nil, fmt.Errorf("package %s: not found", n)
	}

	switch t {
	case Any:
		return newPackage(n, append(c.UserDB, c.SysDB)...)
	case Sys:
		return newPackage(n, c.SysDB)
	case Usr:
		return newPackage(n, c.UserDB...)
	}

	panic("unreachable")
}

// NewPackageByPath finds a package that contains
// given path and returns a pointer to Package.
func NewPackageByPath(c *Config, p string) (*Package, error) {
	err := pathsOnce.Do(func() error {
		add := func(n string) error {
			sp, err := NewPackageByName(c, Sys, n)

			if err != nil {
				return err
			}

			f, err := os.Open(filepath.Join(sp.Path, "manifest"))

			if err != nil {
				return err
			}

			defer f.Close()

			sc := bufio.NewScanner(f)

			for sc.Scan() {
				p := sc.Text()

				if strings.HasSuffix(p, "/") {
					continue
				}

				if op, ok := paths[p]; ok {
					return fmt.Errorf("package %s: owned by %s, %s", p, op.Name, sp.Name)
				}

				paths[p] = sp
			}

			return sc.Err()
		}

		dd, err := os.ReadDir(c.SysDB)

		if err != nil {
			return err
		}

		paths = make(map[string]*Package)

		for _, de := range dd {
			if err := add(de.Name()); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if p, ok := paths[p]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("package %s: not owned", p)
}

func (p *Package) String() string {
	return p.Name
}
