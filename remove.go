package king

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/skel"
)

func (p *Package) Remove(force bool) error {
	pp, err := skel.Slice(filepath.Join(p.Path, "manifest"))

	if err != nil {
		return err
	}

	ee, err := skel.Map(filepath.Join(p.Path, "etcsums"))

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// TODO redo
	if !force {
		rpp, err := p.ReverseDepends()

		if err != nil {
			return err
		}

		if len(rpp) > 0 {
			return fmt.Errorf("package %s required by other packages: %s", p.Name, rpp)
		}
	}

	if err := p.cfg.RunRepoHook("pre-remove", p.Name); err != nil {
		return err
	}

	if err := p.cfg.RunUserHook("pre-remove", p.Name, p.Path); err != nil {
		return err
	}

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	for _, r := range pp {
		rp := filepath.Join(p.cfg.RootDir, r)
		st, err := os.Lstat(rp)

		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return err
		}

		switch m := st.Mode(); {
		case m.IsRegular() && strings.HasPrefix(r, "/etc/"):
			h, err := file.Sha256Sum(rp)

			if err != nil {
				return err
			}

			if len(ee) > 0 && !ee[h] {
				continue
			}
		case m.IsDir():
			dd, err := file.ReadDirNames(rp)

			if err != nil {
				return err
			}

			if len(dd) > 0 {
				continue
			}
		}

		if err := os.Remove(rp); err != nil {
			return err
		}
	}

	signal.Reset(os.Interrupt)
	return p.cfg.RunUserHook("post-remove", p.Name, "null")
}
