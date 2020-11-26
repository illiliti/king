package king

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/file"
	"go4.org/syncutil"
)

var (
	ownerOnce syncutil.Once
	owner     map[string]*Package // TODO synchronize caching with alternative system, remove/install
)

// TODO rename 'Owner' ?
func (c *Config) Owner(p string) (*Package, error) {
	err := ownerOnce.Do(func() error {
		dd, err := file.ReadDirNames(c.SysDB)

		if err != nil {
			return err
		}

		populate := func(n string) error {
			p, err := c.NewPackage(n, Sys)

			if err != nil {
				return err
			}

			f, err := os.Open(filepath.Join(p.Path, "manifest"))

			if err != nil {
				return err
			}

			defer f.Close()

			sc := bufio.NewScanner(f)

			// TODO what if two or more packages owns same path ? this should be considered as bug ?
			for sc.Scan() {
				owner[sc.Text()] = p
			}

			return sc.Err()
		}

		owner = make(map[string]*Package)

		for _, n := range dd {
			if err := populate(n); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if p, ok := owner[p]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("file is not owned: %s", p)
}
