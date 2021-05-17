package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/dustin/go-humanize"
	"github.com/go-git/go-git/v5"
	"github.com/illiliti/king"
	"github.com/illiliti/king/manifest"
)

func query(c *king.Config, args []string) error {
	// TODO better var names
	// TODO add "tree" feature
	var (
		fR bool
		fS bool
		fl bool
		fL bool
		fA bool
		fs bool
		f1 bool
		fz string
		ff string
		fm string
		fd string
		fD string
		fo string
	)

	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	pf.BoolVarP(&fR, "only-repositories", "R", false, "")
	pf.BoolVarP(&fS, "only-database", "S", false, "")
	pf.BoolVarP(&fl, "packages", "l", false, "")
	pf.BoolVarP(&fL, "repositories", "L", false, "")
	pf.BoolVarP(&fA, "alternatives", "A", false, "")
	pf.BoolVarP(&fs, "search", "s", false, "")
	pf.BoolVarP(&f1, "single", "1", false, "")
	pf.StringVarP(&fz, "size", "z", "", "")
	pf.StringVarP(&ff, "files", "f", "", "")
	pf.StringVarP(&fm, "maintainer", "m", "", "")
	pf.StringVarP(&fd, "deps", "d", "", "")
	pf.StringVarP(&fD, "revdeps", "D", "", "")
	pf.StringVarP(&fo, "owner", "o", "", "")

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, queryUsage)
	}

	pf.Parse(args[1:])

	var rt king.RepositoryType

	switch {
	case fR:
		rt = king.Repository
	case fS:
		rt = king.Database
	default:
		rt = king.All
	}

	var err error

	w := bufio.NewWriter(os.Stdout)

	switch {
	case fl:
		err = listPackages(c, w, pf.Args())
	case fL:
		err = listRepositories(c, w) // TODO allow args
	case fA:
		err = listAlternatives(c, w, pf.Args())
	case fs:
		err = searchPackages(c, w, pf.Args(), rt, f1)
	case pf.NArg() > 0:
		pf.Usage()
		os.Exit(2)
	case fz != "":
		err = querySize(c, w, fz)
	case ff != "":
		err = queryManifest(c, w, ff)
	case fm != "":
		err = queryMaintainer(c, w, fm)
	case fd != "":
		err = queryDependencies(c, w, fd, rt)
	case fD != "":
		err = queryReverseDependencies(c, w, fD)
	case fo != "":
		err = queryOwner(c, w, fo)
	default:
		pf.Usage()
		os.Exit(2)
	}

	if err != nil {
		return err
	}

	defer w.Flush()
	return nil
}

// TODO allow king.RepositoryType
func listPackages(c *king.Config, w io.Writer, args []string) error {
	if len(args) == 0 {
		f, err := os.Open(c.DatabaseDir)

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		// TODO sort
		args, err = f.Readdirnames(0)

		if err != nil {
			return err
		}
	}

	for _, n := range args {
		p, err := king.NewPackage(c, &king.PackageOptions{
			Name: n,
			From: king.Database,
		})

		if err != nil {
			return err
		}

		v, err := p.Version()

		if err != nil {
			return err
		}

		fmt.Fprintln(w, p.Name, v.Version, v.Release)
	}

	return nil
}

func listRepositories(c *king.Config, w io.Writer) error {
	if len(c.Repositories) == 0 {
		return errors.New("repositories are unset")
	}

	for _, d := range c.Repositories {
		dd, err := os.ReadDir(d)

		if err != nil {
			return err
		}

		fmt.Fprintln(w, d, len(dd))
	}

	return nil
}

func listAlternatives(c *king.Config, w io.Writer, args []string) error {
	pp := make(map[string]bool, len(args))

	for _, n := range args {
		pp[n] = true
	}

	aa, err := king.Alternatives(c)

	if err != nil {
		return err
	}

	// TODO match by path
	for _, a := range aa {
		if pp == nil || pp[a.Name] {
			fmt.Fprintln(w, a.Name, a.Path)
		}
	}

	return nil
}

func searchPackages(c *king.Config, w io.Writer, args []string, rt king.RepositoryType, s bool) error {
	var rr []string

	switch rt {
	case king.Repository:
		rr = c.Repositories
	case king.Database:
		rr = []string{c.DatabaseDir}
	default:
		rr = append(c.Repositories, c.DatabaseDir)
	}

	for _, n := range args {
		if n == "" {
			return errors.New("arguments must be non-empty")
		}

		for _, r := range rr {
			gg, err := filepath.Glob(filepath.Join(r, n))

			if err != nil {
				log.Fatal(err)
			}

			for _, p := range gg {
				fmt.Fprintln(w, p)

				if s {
					return nil
				}
			}
		}
	}

	return nil
}

func querySize(c *king.Config, w io.Writer, n string) error {
	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: n,
		From: king.Database,
	})

	if err != nil {
		return err
	}

	mf, err := manifest.Open(filepath.Join(p.Path, "manifest"), os.O_RDONLY)

	if err != nil {
		return err
	}

	defer mf.Close()

	var t uint64

	for _, p := range mf.Sort(manifest.Files) {
		// TODO print how many files contains directory ?
		if strings.HasSuffix(p, "/") {
			continue
		}

		st, err := os.Lstat(filepath.Join(c.RootDir, p))

		if err != nil {
			return err
		}

		// TODO warn
		if !st.Mode().IsRegular() {
			continue
		}

		z := uint64(st.Size())
		t += z

		fmt.Fprintln(w, p, humanize.Bytes(z))
	}

	fmt.Fprintf(w, "total %s\n", humanize.Bytes(t))
	return nil
}

func queryManifest(c *king.Config, w io.Writer, n string) error {
	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: n,
		From: king.Database,
	})

	if err != nil {
		return err
	}

	mf, err := manifest.Open(filepath.Join(p.Path, "manifest"), os.O_RDONLY)

	if err != nil {
		return err
	}

	defer mf.Close()

	for _, p := range mf.Sort(manifest.Files) {
		fmt.Fprintln(w, p)
	}

	return nil
}

// TODO move to library?
func queryMaintainer(c *king.Config, w io.Writer, n string) error {
	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: n,
		From: king.Repository,
	})

	if err != nil {
		return err
	}

	rp, err := filepath.EvalSymlinks(p.Path)

	if err != nil {
		return err
	}

	r, err := git.PlainOpenWithOptions(rp, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return err
	}

	// FIXME extremely slow
	// king: king q -m busybox  3.00s user 0.48s system 111% cpu 3.116 total
	// kiss: kiss-maintainer busybox  0.02s user 0.01s system 99% cpu 0.033 total
	ci, err := r.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
		PathFilter: func(gp string) bool {
			if filepath.Base(gp) != "version" {
				return false
			}

			return strings.HasSuffix(rp, filepath.Dir(gp))
		},
	})

	if err != nil {
		return err
	}

	gc, err := ci.Next()

	if err != nil {
		return err
	}

	defer ci.Close()

	fmt.Fprintln(w, gc.Author.String())
	return nil
}

// TODO recursive
func queryDependencies(c *king.Config, w io.Writer, n string, rt king.RepositoryType) error {
	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: n,
		From: rt,
	})

	if err != nil {
		return err
	}

	dd, err := p.Dependencies()
	// dd, err := p.RecursiveDependencies()

	if err != nil {
		return err
	}

	for _, d := range dd {
		fmt.Fprintln(w, d.Name)
	}

	return nil
}

func queryReverseDependencies(c *king.Config, w io.Writer, n string) error {
	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: n,
		// From: rt, // TODO
	})

	if err != nil {
		return err
	}

	dd, err := p.ReverseDependencies()

	if err != nil {
		return err
	}

	for _, d := range dd {
		fmt.Fprintln(w, d)
	}

	return nil
}

func queryOwner(c *king.Config, w io.Writer, p string) error {
	rp, err := filepath.EvalSymlinks(p)

	if err != nil {
		return err
	}

	sp, err := king.NewPackage(c, &king.PackageOptions{
		Path: rp,
	})

	if err != nil {
		return err
	}

	fmt.Fprintln(w, sp.Name)
	return nil
}
