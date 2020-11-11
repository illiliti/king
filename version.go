package king

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Version struct {
	Current string
	Release string
}

func (p *Package) Version() (*Version, error) {
	err := p.versionOnce.Do(func() error {
		b, err := ioutil.ReadFile(filepath.Join(p.Path, "version"))

		if err != nil {
			return err
		}

		vr := strings.Fields(string(b))

		if len(vr) < 2 {
			return fmt.Errorf("invalid version: %s", b)
		}

		p.version = &Version{
			Current: vr[0],
			Release: vr[1],
		}

		return nil
	})

	return p.version, err
}
