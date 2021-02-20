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

var (
	dependenciesCount uint32
	dependenciesMutex sync.Mutex
	dependencies      map[string][]string
)

// Dependency represents a line of the depends file.
//
// See https://kiss.armaanb.net/package-system#3.0
type Dependency struct {
	Name   string
	IsMake bool
}

// Depends returns slice of Dependency pointers for a given package.
func (p *Package) Depends() ([]*Dependency, error) {
	f, err := os.Open(filepath.Join(p.Path, "depends"))

	if err != nil {
		return nil, err
	}

	defer f.Close()

	var dd []*Dependency

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		fi := strings.Fields(sc.Text())

		if len(fi) == 0 || fi[0][0] == '#' {
			continue
		}

		dd = append(dd, &Dependency{
			Name:   fi[0],
			IsMake: len(fi) > 1 && fi[1] == "make",
		})
	}

	return dd, sc.Err()
}

// RecursiveDepends is a recursive version of Depends() which returns
// slice of Dependency pointers.
func (p *Package) RecursiveDepends() ([]*Dependency, error) {
	dd, err := p.Depends()

	if err != nil {
		return nil, err
	}

	for _, d := range dd {
		rp, err := NewPackageByName(p.cfg, Any, d.Name)

		if err != nil {
			return nil, err
		}

		rdd, err := rp.RecursiveDepends()

		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, err
		}

		dd = append(rdd, dd...)
	}

	return dd, nil
}

// ReverseDepends is a reverse version of Depends() that iterates
// over SysDB and returns slice of strings with no make dependencies.
//
// TODO allow UserDB?
func (p *Package) ReverseDepends() ([]string, error) {
	if err := initDependencies(p); err != nil {
		return nil, err
	}

	if dd, ok := dependencies[p.Name]; ok {
		return dd, nil
	}

	return nil, fmt.Errorf("depends %s: no reverse dependencies", p.Name)
}

func initDependencies(p *Package) error {
	if atomic.LoadUint32(&dependenciesCount) == 1 {
		return nil
	}

	dependenciesMutex.Lock()
	defer dependenciesMutex.Unlock()

	if atomic.LoadUint32(&dependenciesCount) == 1 {
		return nil
	}

	dd, err := os.ReadDir(p.cfg.SysDB)

	if err != nil {
		return err
	}

	dependencies = make(map[string][]string, len(dd))

	for _, de := range dd {
		sp, err := NewPackageByName(p.cfg, Sys, de.Name())

		if err != nil {
			return err
		}

		dd, err := sp.Depends()

		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return err
		}

		for _, d := range dd {
			if d.IsMake {
				continue
			}

			dependencies[d.Name] = append(dependencies[d.Name], sp.Name)
		}
	}

	atomic.StoreUint32(&dependenciesCount, 1)
	return nil
}
