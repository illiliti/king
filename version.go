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
	ErrVersionInvalid = errors.New("target must contain only two fields")
)

// Version represents content of the version file.
//
// See https://k1sslinux.org/package-system#5.0
type Version struct {
	Version string
	Release string
}

func (p *Package) Version() (*Version, error) {
	b, err := os.ReadFile(filepath.Join(p.Path, "version"))

	if err != nil {
		return nil, err
	}

	vr := strings.Fields(string(b))

	if len(vr) == 2 {
		return &Version{
			Version: vr[0],
			// TODO must be uint
			// TODO must be > 0
			Release: vr[1],
		}, nil
	}

	return nil, fmt.Errorf("parse %s version: %w", p.Name, ErrVersionInvalid)
}

func (v *Version) String() string {
	return v.Version + " " + v.Release
}
