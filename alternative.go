package king

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/skel/manifest"
)

type Alternative struct {
	Name string
	Path string

	cfg *Config
}

func (c *Config) NewAlternative(n, p string) (*Alternative, error) {
	_, err := os.Lstat(filepath.Join(c.RootDir, ChoicesDir,
		n+strings.ReplaceAll(p, "/", ">")))

	return &Alternative{
		Name: n,
		Path: p,
		cfg:  c,
	}, err
}

func (c *Config) Alternatives() ([]*Alternative, error) {
	dd, err := file.ReadDirNames(filepath.Join(c.RootDir, ChoicesDir))

	if err != nil {
		return nil, err
	}

	aa := make([]*Alternative, 0, len(dd))

	for _, n := range dd {
		i := strings.Index(n, ">")

		if i < 0 {
			return nil, fmt.Errorf("invalid alternative: %s", n)
		}

		aa = append(aa, &Alternative{
			Name: n[:i],
			Path: strings.ReplaceAll(n[i:], ">", "/"),
			cfg:  c,
		})
	}

	return aa, nil
}

func (a *Alternative) Swap() (*Alternative, error) {
	sp, err := a.cfg.NewPackage(a.Name, Sys)

	if err != nil {
		return nil, err
	}

	ap := strings.ReplaceAll(a.Path, "/", ">")
	rp := filepath.Join(a.cfg.RootDir, a.Path)
	cp, err := a.cfg.Owner(a.Path)

	if err != nil {
		return nil, err
	}

	cn := cp.Name + ap

	// TODO what if package removed ?
	if err := os.Rename(rp, filepath.Join(a.cfg.RootDir, ChoicesDir, cn)); err != nil {
		return nil, err
	}

	if err := manifest.Replace(filepath.Join(cp.Path, "manifest"), a.Path, filepath.Join(ChoicesDir, cn)); err != nil {
		return nil, err
	}

	an := sp.Name + ap

	if err := os.Rename(filepath.Join(a.cfg.RootDir, ChoicesDir, an), rp); err != nil {
		return nil, err
	}

	return &Alternative{
		Name: cp.Name,
		Path: a.Path,
		cfg:  a.cfg,
	}, manifest.Replace(filepath.Join(sp.Path, "manifest"), filepath.Join(ChoicesDir, an), a.Path)
}
