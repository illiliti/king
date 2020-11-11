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
	reverseDepends     map[string][]string
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

			dp := &Dependency{
				Name: fi[0],
			}

			if len(fi) == 2 && fi[1] == "make" {
				dp.IsMake = true
			}

			p.depends = append(p.depends, dp)
		}

		return sc.Err()
	})

	return p.depends, err
}

func (p *Package) RecursiveDepends() ([]*Package, error) {
	pp, err := p.Depends()

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var dpp []*Package

	for _, dp := range pp {
		if _, err := p.context.NewPackage(dp.Name, SysDB); !os.IsNotExist(err) {
			continue
		}

		p, err := p.context.NewPackage(dp.Name, UserDB)

		if err != nil {
			return nil, err
		}

		rpp, err := p.RecursiveDepends()

		if err != nil {
			return nil, err
		}

		dpp = append(dpp, rpp...)
		dpp = append(dpp, p)
	}

	return dpp, nil
}

// TODO UserDB ?
func (p *Package) ReverseDepends() ([]string, error) {
	err := reverseDependsOnce.Do(func() error {
		dd, err := file.ReadDirNames(p.context.SysDB)

		if err != nil {
			return err
		}

		reverseDepends = make(map[string][]string, len(dd))

		for _, n := range dd {
			sp, err := p.context.NewPackage(n, SysDB)

			if err != nil {
				return err
			}

			pp, err := sp.Depends()

			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return err
			}

			for _, dp := range pp {
				if dp.IsMake {
					continue
				}

				reverseDepends[dp.Name] = append(reverseDepends[dp.Name], sp.Name)
			}
		}

		return nil
	})

	return reverseDepends[p.Name], err
}
