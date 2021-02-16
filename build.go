package king

import (
	"bufio"
	"debug/elf"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/etcsum"
	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/manifest"
)

func (p *Package) Build() (*Tarball, error) {
	pd := filepath.Join(p.cfg.PkgDir, p.Name)
	bd := filepath.Join(p.cfg.BuildDir, p.Name)
	pdp := filepath.Join(pd, InstalledDir, p.Name)
	pdl := filepath.Join(pd, "usr/lib")

	v, err := p.Version()

	if err != nil {
		return nil, err
	}

	ss, err := p.Sources()

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
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

	dd, err := os.ReadDir(pdl)

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	for _, de := range dd {
		if de.Name() != "charset.alias" && filepath.Ext(de.Name()) != ".la" {
			continue
		}

		if err := os.Remove(filepath.Join(pdl, de.Name())); err != nil {
			return nil, err
		}
	}

	if err := file.CopyDir(p.Path, pdp); err != nil {
		return nil, err
	}

	if err := updateDepends(p, pd, pdp); err != nil {
		return nil, err
	}

	// TODO strip binaries
	// if err := stripBinaries(pd); err != nil {
	// 	return nil, err
	// }

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

	if err := os.Remove(t.Path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	return t, file.Archive(pd, t.Path)
}

func updateDepends(bp *Package, pd, pdp string) error {
	mdd := make(map[string]bool)

	err := filepath.WalkDir(pd, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !de.Type().IsRegular() {
			return nil
		}

		f, err := elf.Open(p)

		if err != nil {
			return nil
		}

		ll, err := f.ImportedLibraries()

		if err != nil {
			return nil
		}

		for _, l := range ll {
			i := strings.Index(l, ".")

			if i < 0 {
				continue
			}

			switch l[3:i] {
			case "c", "m", "dl", "rt", "xnet", "util", "trace", "crypt", "pthread", "resolv":
				continue
			}

			p, err := bp.cfg.NewPackageByPath(filepath.Join("/usr/lib", l))

			if err != nil {
				continue
			}

			if p.Name == bp.Name {
				continue
			}

			if !mdd[p.Name] {
				mdd[p.Name] = true
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(mdd) == 0 {
		return nil
	}

	f, err := os.OpenFile(filepath.Join(pdp, "depends"), os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		return err
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		d := strings.Fields(sc.Text())[0]

		if !mdd[d] {
			mdd[d] = true
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	if err := f.Truncate(0); err != nil {
		return err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	for n := range mdd {
		fmt.Fprintln(w, n)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return f.Close()
}
