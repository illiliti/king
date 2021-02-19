package king

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Version represents content of the version file.
//
// See https://kiss.armaanb.net/package-system#5.0
type Version struct {
	Version string
	Release string
}

// Version returns a pointer to Version for a given package.
func (p *Package) Version() (*Version, error) {
	b, err := os.ReadFile(filepath.Join(p.Path, "version"))

	if err != nil {
		return nil, err
	}

	vr := strings.Fields(string(b))

	if len(vr) != 2 {
		return nil, fmt.Errorf("version %s: malformed", b)
	}

	return &Version{
		Version: vr[0],
		Release: vr[1],
	}, nil
}
