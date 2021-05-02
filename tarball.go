package king

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TODO better docs

var (
	ErrTarballNotFound   = errors.New("target has not been built yet")
	ErrTarballNotRegular = errors.New("target must be a regular file")
	ErrTarballInvalid    = errors.New("target has no separator")
)

// Tarball represents pre-built package.
type Tarball struct {
	Name string
	Path string

	cfg *Config
}

// NewTarball allocates new instance of pre-built package.
func NewTarball(c *Config, p string) (*Tarball, error) {
	st, err := os.Stat(p)

	if err != nil {
		return nil, err
	}

	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("parse tarball %s: %w", p, ErrTarballNotRegular)
	}

	i := strings.Index(p, "@")

	if i < 1 {
		return nil, fmt.Errorf("parse tarball %s: %w", p, ErrTarballInvalid)
	}

	return &Tarball{
		Name: p[:i],
		Path: p,
		cfg:  c,
	}, nil
}

// Tarball tries to find pre-built tarball within BinaryDir.
func (p *Package) Tarball() (*Tarball, error) {
	v, err := p.Version()

	if err != nil {
		return nil, err
	}

	n := p.Name + "@" + v.Version + "-" + v.Release + "*"
	pp, err := filepath.Glob(filepath.Join(p.cfg.BinaryDir, n))

	if err != nil {
		return nil, err
	}

	if len(pp) == 0 {
		return nil, fmt.Errorf("find tarball by name %s: %w", p.Name, ErrTarballNotFound)
	}

	return &Tarball{
		Name: p.Name,
		Path: pp[0],
		cfg:  p.cfg,
	}, nil
}

func (t *Tarball) String() string {
	return t.Path
}
