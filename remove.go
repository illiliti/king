package king

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/file"
)

func (p *Package) Remove(force bool) error {
	pp, err := readManifest(filepath.Join(p.Path, "manifest"))

	if err != nil {
		return err
	}

	ee, err := readEtcsums(filepath.Join(p.Path, "etcsums"))

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if !force {
		rpp, err := p.ReverseDepends()

		if err != nil {
			return err
		}

		if len(rpp) > 0 {
			return fmt.Errorf("package %s required by other packages: %s", p.Name, rpp)
		}
	}

	if err := p.context.RunRepoHook("pre-remove", p.Name); err != nil {
		return err
	}

	if err := p.context.RunUserHook("pre-remove", p.Name, p.Path); err != nil {
		return err
	}

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	for _, r := range pp {
		rp := filepath.Join(p.context.RootDir, r)
		st, err := os.Lstat(rp)

		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return err
		}

		m := st.Mode()

		if m.IsRegular() && strings.HasPrefix(r, "/etc/") {
			h, err := file.Sha256Sum(rp)

			if err != nil {
				return err
			}

			if len(ee) > 0 && !ee[h] {
				continue
			}
		}

		if err := removeFile(rp, m); err != nil {
			return err
		}
	}

	signal.Reset(os.Interrupt)
	return p.context.RunUserHook("post-remove", p.Name, "null")
}
