package king

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// TODO better docs

var (
	ErrPackagePathNotFound = errors.New("target not owned by any package")
	ErrPackageNameNotFound = errors.New("target not found within specified database")
)

type RepositoryType uint

const (
	All RepositoryType = iota
	Database
	Repository
)

// Package represents package within repository or database.
//
// See https://k1sslinux.org/package-system#1.0
type Package struct {
	Name string
	Path string

	From RepositoryType

	cfg *Config
}

// PackageOptions intended to configure searching of package.
type PackageOptions struct {
	// Name defines package name that will be used to traverse
	// repositories or database.
	Name string

	// Path defines path that will be used to grep manifests files
	// within database to find owner of that path.
	Path string

	// From defines where to search package. Applies only to Name field.
	From RepositoryType
}

// NewPackage allocates new instance of package.
func NewPackage(c *Config, po *PackageOptions) (*Package, error) {
	if err := po.Validate(); err != nil {
		return nil, fmt.Errorf("validate PackageOptions: %w", err)
	}

	if po.Path != "" {
		return newPackageByPath(c, po.Path)
	}

	switch po.From {
	case All:
		return newPackageByName(c, po, append(c.Repositories, c.DatabaseDir)...)
	case Database:
		return newPackageByName(c, po, c.DatabaseDir)
	case Repository:
		return newPackageByName(c, po, c.Repositories...)
	}

	panic("unreachable")
}

// TODO check essential files (build, manifest, ...)
func newPackageByName(c *Config, po *PackageOptions, dd ...string) (*Package, error) {
	for _, d := range dd {
		p := filepath.Join(d, po.Name)
		st, err := os.Stat(p)

		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, err
		}

		// TODO panic()
		if !st.IsDir() {
			continue
		}

		return &Package{
			Name: po.Name,
			Path: p,
			From: po.From,
			cfg:  c,
		}, nil
	}

	return nil, fmt.Errorf("find package by name %s: %w", po.Name, ErrPackageNameNotFound)
}

// TODO fail if directory
func newPackageByPath(c *Config, p string) (*Package, error) {
	if err := c.initOwnedPaths(); err != nil {
		return nil, fmt.Errorf("initialize owned paths: %w", err)
	}

	// c.ppm.Lock()
	// defer c.ppm.Unlock()

	if p, ok := c.pp[p]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("find package by path %s: %w", p, ErrPackagePathNotFound)
}

func (rt RepositoryType) String() string {
	switch rt {
	case Repository:
		return "repository"
	case Database:
		return "database"
	case All:
		return "all"
	}

	panic("unreachable")
}

func (p *Package) String() string {
	return p.Name
}
