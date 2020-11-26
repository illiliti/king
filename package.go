package king

import (
	"fmt"
	"os"
	"path/filepath"

	"go4.org/syncutil"
)

type Package struct {
	Name string
	Path string

	cfg *Config

	checksumsOnce syncutil.Once
	checksums     map[string]bool

	dependsOnce syncutil.Once
	depends     []*Dependency

	sourcesOnce syncutil.Once
	sources     []*Source

	versionOnce syncutil.Once
	version     *Version
}

type PackageType int

const (
	Any PackageType = iota
	Sys
	User
)

func (c *Config) NewPackage(n string, t PackageType) (*Package, error) {
	findPackage := func(n string, dd ...string) (*Package, error) {
		for _, db := range dd {
			p := filepath.Join(db, n)
			st, err := os.Stat(p)

			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return nil, err
			}

			if !st.IsDir() {
				return nil, fmt.Errorf("invalid package: %s", n)
			}

			return &Package{
				Name: n,
				Path: p,
				cfg:  c,
			}, nil
		}

		return nil, fmt.Errorf("package not found: %s", n)
	}

	switch t {
	case Any:
		return findPackage(n, append(c.UserDB, c.SysDB)...)
	case Sys:
		return findPackage(n, c.SysDB)
	case User:
		return findPackage(n, c.UserDB...)
	}

	panic("unreachable")
}
