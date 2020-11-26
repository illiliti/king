package king

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/file"
	"go4.org/syncutil"
)

var (
	reverseDependsOnce syncutil.Once
	reverseDepends     map[string][]*Package
)

type Dependency struct {
	Name   string
	IsMake bool
}

func (p *Package) Depends() ([]*Dependency, error) {
	err := p.dependsOnce.Do(func() error {
		f, err := os.Open(filepath.Join(p.Path, "depends"))

		if err != nil {
			return err
		}

		defer f.Close()

		sc := bufio.NewScanner(f)

		for sc.Scan() {
			fi := strings.Fields(sc.Text())

			if len(fi) == 0 || fi[0][0] == '#' {
				continue
			}

			d := &Dependency{
				Name: fi[0],
			}

			if len(fi) == 2 && fi[1] == "make" {
				d.IsMake = true
			}

			p.depends = append(p.depends, d)
		}

		return sc.Err()
	})

	return p.depends, err
}

func (p *Package) RecursiveDepends() ([]*Dependency, error) {
	dd, err := p.Depends()

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	for _, d := range dd {
		rp, err := p.cfg.NewPackage(d.Name, Any)

		if err != nil {
			return nil, err
		}

		rdd, err := rp.RecursiveDepends()

		if err != nil {
			return nil, err
		}

		dd = append(rdd, dd...)
	}

	return dd, nil
}

// TODO UserDB ?
func (p *Package) ReverseDepends() ([]*Package, error) {
	err := reverseDependsOnce.Do(func() error {
		dd, err := file.ReadDirNames(p.cfg.SysDB)

		if err != nil {
			return err
		}

		reverseDepends = make(map[string][]*Package, len(dd))

		for _, n := range dd {
			sp, err := p.cfg.NewPackage(n, Sys)

			if err != nil {
				return err
			}

			dd, err := sp.Depends()

			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return err
			}

			for _, d := range dd {
				if d.IsMake {
					continue
				}

				reverseDepends[d.Name] = append(reverseDepends[d.Name], sp)
			}
		}

		return nil
	})

	return reverseDepends[p.Name], err
}
