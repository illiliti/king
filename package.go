package king

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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
	pathsCount uint32
	pathsMutex sync.Mutex
	paths      map[string]*Package
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
	switch t {
	case Any:
		return newPackage(c, n, append(c.UserDB, c.SysDB)...)
	case Sys:
		return newPackage(c, n, c.SysDB)
	case Usr:
		return newPackage(c, n, c.UserDB...)
	}

	panic("unreachable")
}

func newPackage(c *Config, n string, dd ...string) (*Package, error) {
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

// NewPackageByPath finds a package that contains
// given path and returns a pointer to Package.
func NewPackageByPath(c *Config, p string) (*Package, error) {
	if err := initPaths(c); err != nil {
		return nil, err
	}

	if p, ok := paths[p]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("package %s: not owned", p)
}

func initPaths(c *Config) error {
	if atomic.LoadUint32(&pathsCount) == 1 {
		return nil
	}

	pathsMutex.Lock()
	defer pathsMutex.Unlock()

	if atomic.LoadUint32(&pathsCount) == 1 {
		return nil
	}

	dd, err := os.ReadDir(c.SysDB)

	if err != nil {
		return err
	}

	paths = make(map[string]*Package)

	for _, de := range dd {
		sp, err := NewPackageByName(c, Sys, de.Name())

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

		if err := sc.Err(); err != nil {
			return err
		}
	}

	atomic.StoreUint32(&pathsCount, 1)
	return nil
}

func (p *Package) String() string {
	return p.Name
}
