package king

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/skel"
	"github.com/illiliti/king/internal/skel/etcsums"
	"github.com/illiliti/king/internal/skel/manifest"
)

// TODO cleanup on signal
func (p *Package) Build() (*Tarball, error) {
	bd := filepath.Join(p.cfg.BuildDir, p.Name)
	pd := filepath.Join(p.cfg.PkgDir, p.Name)

	for _, d := range []string{p.cfg.LogDir, bd, pd} {
		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}
	}

	if !p.cfg.HasDebug {
		defer func() {
			for _, d := range []string{bd, pd} {
				os.RemoveAll(d)
			}
		}()
	}

	if err := p.cfg.RunUserHook("pre-extract", p.Name, pd); err != nil {
		return nil, err
	}

	ss, err := p.Sources()

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	for _, s := range ss {
		// if d, ok := s.Protocol.(Downloader); ok {
		// 	if err := d.Download(); err != nil {
		// 		return nil, err
		// 	}
		// }

		// if v, ok := s.Protocol.(Verifier); ok {
		// 	if err := v.Verify(); err != nil {
		// 		return nil, err
		// 	}
		// }

		p, ok := s.Protocol.(Preparer)

		if !ok {
			continue
		}

		d := filepath.Join(bd, s.CustomDir)

		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}

		if err := p.Prepare(d); err != nil {
			return nil, err
		}
	}

	if err := p.cfg.RunUserHook("pre-build", p.Name, pd); err != nil {
		return nil, err
	}

	v, err := p.Version()

	if err != nil {
		return nil, err
	}

	l := filepath.Join(p.cfg.LogDir, p.Name+"-"+
		time.Now().Format("2006-01-02-15:04")+"-"+p.cfg.ProcessID)

	f, err := os.Create(l)

	if err != nil {
		return nil, err
	}

	w := io.MultiWriter(os.Stdout, f)

	cmd := exec.Command(filepath.Join(p.Path, "build"), pd, v.Current)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Dir = bd

	if err := cmd.Run(); err != nil {
		if err := p.cfg.RunUserHook("build-fail", p.Name, pd); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	if !p.cfg.HasKeepLog {
		if err := os.Remove(l); err != nil {
			return nil, err
		}
	}

	if err := p.cfg.RunUserHook("post-build", p.Name, pd); err != nil {
		return nil, err
	}

	// TODO remove ".la" and "charset.alias"
	// TODO stripBinaries
	// TODO correctDepends

	pdp := filepath.Join(pd, InstalledDir, p.Name)

	if err := file.CopyDir(p.Path, pdp); err != nil {
		return nil, err
	}

	ee, err := etcsums.Generate(pd)

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(ee) > 0 {
		f, err := os.Create(filepath.Join(pdp, "etcsums"))

		if err != nil {
			return nil, err
		}

		if err := skel.Save(ee, f); err != nil {
			return nil, err
		}

		if err := f.Close(); err != nil {
			return nil, err
		}
	}

	f, err = os.Create(filepath.Join(pdp, "manifest"))

	if err != nil {
		return nil, err
	}

	pp, err := manifest.Generate(pd)

	if err != nil {
		return nil, err
	}

	if err := skel.Save(pp, f); err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	t := &Tarball{
		Name: p.Name,
		Path: filepath.Join(p.cfg.BinDir, p.Name+"@"+v.Current+
			"-"+v.Release+".tar."+p.cfg.CompressFormat),
		cfg: p.cfg,
	}

	if err := os.Remove(t.Path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := file.CreateArchive(pd, t.Path); err != nil {
		return nil, err
	}

	return t, p.cfg.RunUserHook("post-package", p.Name, "null")
}
