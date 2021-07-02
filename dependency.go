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

var ErrReverseDependenciesNotFound = errors.New("no package depends on target")

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

func (p *Package) ReverseDependencies() ([]string, error) {
	if err := p.cfg.initReverseDependencies(); err != nil {
		return nil, fmt.Errorf("initialize reverse dependencies: %w", err)
	}

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
