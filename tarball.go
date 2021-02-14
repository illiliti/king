package king

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TODO String() ?
type Tarball struct {
	Name string
	Path string

	cfg *Config
}

func (p *Package) Tarball() (*Tarball, error) {
	v, err := p.Version()

	if err != nil {
		return nil, err
	}

	t := p.Name + "[@#]" + v.Version + "-" + v.Release + "*"
	mm, err := filepath.Glob(filepath.Join(p.cfg.BinDir, t))

	if err != nil {
		return nil, err
	}

	if len(mm) == 0 {
		return nil, fmt.Errorf("tarball %s: not found", p.Name)
	}

	return &Tarball{
		Name: p.Name,
		Path: mm[0],
		cfg:  p.cfg,
	}, nil
}

func (c *Config) NewTarball(p string) (*Tarball, error) {
	st, err := os.Stat(p)

	if err != nil {
		return nil, err
	}

	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("tarball %s: not a regular file", p)
	}

	s := strings.IndexAny(p, "@#")

	if s < 0 {
		return nil, fmt.Errorf("tarball %s: missing separator", p)
	}

	return &Tarball{
		Name: p[:s],
		Path: p,
		cfg:  c,
	}, nil
}
