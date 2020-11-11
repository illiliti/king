package king

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/illiliti/king/internal/fetch"
	"github.com/illiliti/king/internal/file"
)

// TODO signal handling and cleanup
func (p *Package) Build() (*Tarball, error) {
	ss, err := p.Sources()

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	v, err := p.Version()

	if err != nil {
		return nil, err
	}

	bd := filepath.Join(p.context.BuildDir, p.Name)
	pd := filepath.Join(p.context.PkgDir, p.Name)

	for _, d := range []string{p.context.LogDir, bd, pd} {
		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}
	}

	prepareSource := func(p interface{}, d string) error {
		switch v := p.(type) {
		case *Git:
			// TODO https://github.com/go-git/go-git/pull/58
			if strings.ContainsAny(v.URL, "#@") {
				return fmt.Errorf("branch or commit is not yet supported: %s", v.URL)
			}

			return fetch.GitClone(v.URL, d)
		case *HTTP:
			if v.HasNoExtract {
				return file.CopyRegular(v.Path, d)
			}

			return file.ExtractArchive(v.Path, d, 1)
		case *File:
			if v.IsDir {
				return file.CopyDir(v.Path, d)
			}

			return file.CopyRegular(v.Path, d)
		}

		panic("unreachable")
	}

	if err := p.context.RunUserHook("pre-extract", p.Name, pd); err != nil {
		return nil, err
	}

	for _, s := range ss {
		d := filepath.Join(bd, s.CustomDir)

		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}

		if err := prepareSource(s.Protocol, d); err != nil {
			return nil, err
		}
	}

	l := filepath.Join(p.context.LogDir, p.Name+"-"+
		time.Now().Format("2006-01-02-15:04")+"-"+p.context.ProcessID)

	startBuild := func() error {
		f, err := os.Create(l)

		if err != nil {
			return err
		}

		w := io.MultiWriter(os.Stdout, f)

		cmd := exec.Command(filepath.Join(p.Path, "build"), pd, v.Current)
		cmd.Stderr = w
		cmd.Stdout = w
		cmd.Dir = bd

		if err := cmd.Run(); err != nil {
			return err
		}

		return f.Close()
	}

	if err := p.context.RunUserHook("pre-build", p.Name, pd); err != nil {
		return nil, err
	}

	if err := startBuild(); err != nil {
		if err := p.context.RunUserHook("build-fail", p.Name, pd); err != nil {
			return nil, err
		}

		return nil, err
	}

	if err := p.context.RunUserHook("post-build", p.Name, pd); err != nil {
		return nil, err
	}

	if !p.context.HasKeepLog {
		if err := os.Remove(l); err != nil {
			return nil, err
		}
	}

	removeGarbage := func() error {
		pdl := filepath.Join(pd, "lib")
		dd, err := file.ReadDirNames(pdl)

		if err != nil {
			return err
		}

		for _, n := range dd {
			if filepath.Ext(n) != ".la" {
				continue
			}

			if err := os.Remove(filepath.Join(pdl, n)); err != nil {
				return err
			}
		}

		return os.Remove(filepath.Join(pdl, "charset.alias"))
	}

	if err := removeGarbage(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// TODO stripBinaries
	// TODO correctDepends

	generateSkeleton := func() error {
		pdp := filepath.Join(pd, "var/db/kiss/installed", p.Name)

		if err := file.CopyDir(p.Path, pdp); err != nil {
			return err
		}

		ee, err := generateEtcsums(pd)

		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if len(ee) > 0 {
			if err := saveEtcsums(ee, filepath.Join(pdp, "etcsums")); err != nil {
				return err
			}
		}

		pp, err := generateManifest(pd)

		if err != nil {
			return err
		}

		return saveManifest(pp, filepath.Join(pdp, "manifest"))
	}

	if err := generateSkeleton(); err != nil {
		return nil, err
	}

	t := &Tarball{
		Name: p.Name,
		Path: filepath.Join(p.context.BinDir, p.Name+"@"+v.Current+
			"-"+v.Release+".tar."+p.context.CompressFormat),
		context: p.context,
	}

	if err := os.Remove(t.Path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := file.CreateArchive(pd, t.Path); err != nil {
		return nil, err
	}

	return t, p.context.RunUserHook("post-package", p.Name, "null")
}
