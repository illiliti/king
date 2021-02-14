package king

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/once"
)

type PackageType uint

const (
	Any PackageType = iota
	Sys
	Usr
)

var (
	paths     map[string]*Package
	pathsOnce once.Once
)

// TODO String() ?
type Package struct {
	Name string
	Path string

	cfg *Config
}

func (c *Config) NewPackageByName(t PackageType, n string) (*Package, error) {
	newPackage := func(n string, dd ...string) (*Package, error) {
		for _, db := range dd {
			p := filepath.Join(db, n)
			st, err := os.Stat(p)

			if os.IsNotExist(err) {
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

func (c *Config) NewPackageByPath(p string) (*Package, error) {
	err := pathsOnce.Do(func() error {
		add := func(n string) error {
			sp, err := c.NewPackageByName(Sys, n)

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

		dd, err := file.ReadDirNames(c.SysDB)

		if err != nil {
			return err
		}

		paths = make(map[string]*Package)

		for _, n := range dd {
			if err := add(n); err != nil {
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
