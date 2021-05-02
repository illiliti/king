package king

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TODO better docs

var (
	ErrReverseDependenciesNotFound = errors.New("no package depends on target")
)

// Dependency represents dependency of package.
//
// See https://k1sslinux.org/package-system#3.0
type Dependency struct {
	Name   string
	IsMake bool
}

func (p *Package) Dependencies() ([]*Dependency, error) {
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

		if fi[0] == p.Name {
			panic(fmt.Sprintf("parse %s dependencies: target depends on itself", p.Name))
		}

		dd = append(dd, &Dependency{
			Name:   fi[0],
			IsMake: len(fi) > 1 && fi[1] == "make",
		})
	}

	return dd, sc.Err()
}

func (p *Package) RecursiveDependencies() ([]*Dependency, error) {
	return recursiveDependencies(p, "", make(map[string]bool))
}

func recursiveDependencies(p *Package, n string, mdd map[string]bool) ([]*Dependency, error) {
	dd, err := p.Dependencies()

	if err != nil {
		return nil, err
	}

	dpp := make([]*Package, 0, len(dd))

	for _, d := range dd {
		if d.Name == n {
			panic(fmt.Sprintf("parse %s dependencies: circular dependencies: [%s %s]", p.Name, d.Name, n))
		}

		if mdd[d.Name] {
			continue
		}

		mdd[d.Name] = true

		dp, err := NewPackage(p.cfg, &PackageOptions{
			Name: d.Name,
			From: p.From,
		})

		if err != nil {
			return nil, err
		}

		dpp = append(dpp, dp)
	}

	for _, dp := range dpp {
		rdd, err := recursiveDependencies(dp, p.Name, mdd)

		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, err
		}

		sdd := make([]*Dependency, 0, len(rdd))

		for _, d := range rdd {
			if mdd[d.Name] {
				continue
			}

			mdd[d.Name] = true
			sdd = append(sdd, d)
		}

		dd = append(sdd, dd...)
	}

	return dd, nil
}

func (p *Package) ReverseDependencies() ([]string, error) {
	if err := p.cfg.initReverseDependencies(); err != nil {
		return nil, fmt.Errorf("initialize reverse dependencies: %w", err)
	}

	// c.ddm.Lock()
	// defer c.ddm.Unlock()

	if dd, ok := p.cfg.dd[p.Name]; ok {
		return dd, nil
	}

	return nil, fmt.Errorf("parse %s reverse dependencies: %w", p.Name, ErrReverseDependenciesNotFound)
}

func (d *Dependency) String() string {
	s := d.Name

	if d.IsMake {
		s += " make"
	}

	return s
}
