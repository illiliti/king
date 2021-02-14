package king

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/illiliti/king/internal/etcsum"
	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/manifest"
)

func (p *Package) Build() (*Tarball, error) {
	pd := filepath.Join(p.cfg.PkgDir, p.Name)
	bd := filepath.Join(p.cfg.BuildDir, p.Name)
	pdp := filepath.Join(pd, InstalledDir, p.Name)

	v, err := p.Version()

	if err != nil {
		return nil, err
	}

	ss, err := p.Sources()

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	for _, d := range []string{p.cfg.LogDir, bd, pd} {
		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}
	}

	// TODO cleanup on signal
	if !p.cfg.HasDebug {
		defer func() {
			for _, d := range []string{bd, pd} {
				os.RemoveAll(d)
			}
		}()
	}

	for _, s := range ss {
		d := filepath.Join(bd, s.DestinationDir)

		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}

		if err := s.Prepare(d); err != nil {
			return nil, err
		}
	}

	// TODO logging to file and stdout
	cmd := exec.Command(filepath.Join(p.Path, "build"), pd, v.Version)
	cmd.Dir = bd

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// TODO remove ".la" and "charset.alias"
	// TODO stripBinaries (strip)
	// TODO correctDepends (ldd)

	if err := file.CopyDir(p.Path, pdp); err != nil {
		return nil, err
	}

	if st, err := os.Stat(filepath.Join(pd, "etc")); err == nil && st.IsDir() {
		e, err := etcsum.Create(filepath.Join(pdp, "etcsums"))

		if err != nil {
			return nil, err
		}

		if err := e.Generate(filepath.Join(pd, "etc")); err != nil {
			return nil, err
		}

		if err := e.Flush(); err != nil {
			return nil, err
		}

		if err := e.Close(); err != nil {
			return nil, err
		}
	}

	m, err := manifest.Create(filepath.Join(pdp, "manifest"))

	if err != nil {
		return nil, err
	}

	if err := m.Generate(pd); err != nil {
		return nil, err
	}

	if err := m.Flush(); err != nil {
		return nil, err
	}

	if err := m.Close(); err != nil {
		return nil, err
	}

	t := &Tarball{
		Name: p.Name,
		Path: filepath.Join(p.cfg.BinDir, p.Name+"@"+v.Version+
			"-"+v.Release+".tar."+p.cfg.CompressFormat),
		cfg: p.cfg,
	}

	if err := os.Remove(t.Path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return t, file.Archive(pd, t.Path)
}
