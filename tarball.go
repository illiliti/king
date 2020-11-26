package king

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Tarball struct {
	Name string
	Path string

	cfg *Config
}

func (c *Config) Tarball(n string) (*Tarball, error) {
	findTarballByName := func(n string) (*Tarball, error) {
		p, err := c.NewPackage(n, Any)

		if err != nil {
			return nil, err
		}

		v, err := p.Version()

		if err != nil {
			return nil, err
		}

		pt := p.Name + "[@#]" + v.Current + "-" + v.Release + ".tar*"
		tt, _ := filepath.Glob(filepath.Join(c.BinDir, pt))

		if len(tt) == 0 {
			return nil, fmt.Errorf("package is not built yet: %s", p.Name)
		}

		return &Tarball{
			Name: p.Name,
			Path: tt[0],
			cfg:  c,
		}, nil
	}

	findTarballByPath := func(p string, m os.FileMode) (*Tarball, error) {
		n := filepath.Base(p)
		s := strings.IndexAny(n, "@#")

		if !m.IsRegular() || s < 0 {
			return nil, fmt.Errorf("invalid tarball: %s", p)
		}

		return &Tarball{
			Name: n[:s],
			Path: p,
			cfg:  c,
		}, nil
	}

	// TODO prefer package over file if file is invalid ?
	st, err := os.Stat(n)

	if err != nil {
		if os.IsNotExist(err) {
			return findTarballByName(n)
		}

		return nil, err
	}

	return findTarballByPath(n, st.Mode())
}
