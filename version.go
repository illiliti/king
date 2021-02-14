package king

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// TODO String() ?
type Version struct {
	Version string
	Release string
}

func (p *Package) Version() (*Version, error) {
	b, err := ioutil.ReadFile(filepath.Join(p.Path, "version"))

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
