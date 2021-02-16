package king

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/once"
)

var (
	dependenciesOnce once.Once
	dependencies     map[string][]string
)

// TODO String() ?
type Dependency struct {
	Name   string
	IsMake bool
}

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

func (p *Package) RecursiveDepends() ([]*Dependency, error) {
	dd, err := p.Depends()

	if err != nil {
		return nil, err
	}

	for _, d := range dd {
		rp, err := p.cfg.NewPackageByName(Any, d.Name)

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

// TODO allow UserDB ?
func (p *Package) ReverseDepends() ([]string, error) {
	err := dependenciesOnce.Do(func() error {
		dd, err := os.ReadDir(p.cfg.SysDB)

		if err != nil {
			return err
		}

		dependencies = make(map[string][]string, len(dd))

		for _, de := range dd {
			sp, err := p.cfg.NewPackageByName(Sys, de.Name())

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

		return nil
	})

	return dependencies[p.Name], err
}
