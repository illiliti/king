package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/illiliti/king/internal/cleanup"
	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
)

// TODO print how many packages are built/building in terminal title
// TODO print elapsed time

func build(c *king.Config, td string, args []string) error {
	var (
		fs bool
		fi bool
		fd bool
		fT bool
		fq bool
		fn bool
	)

	bo := new(king.BuildOptions)
	lo := new(king.InstallOptions)
	do := new(king.DownloadOptions)

	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	// TODO we need structure-based flag parser at some point to avoid this mess
	pf.BoolVar(&bo.AllowInternet, "who-needs-checksums", false, "")
	pf.StringVarP(&lo.ExtractDir, "extract-dir", "X", filepath.Join(td, "extract"), "")
	pf.StringVarP(&bo.PackageDir, "package-dir", "P", filepath.Join(td, "pkg"), "")
	// pf.StringVarP(&fO, "output-dir", "O", filepath.Join(cd, "logs"), "")
	pf.StringVarP(&bo.BuildDir, "build-dir", "B", filepath.Join(td, "build"), "")
	pf.StringVarP(&bo.Compression, "compression", "C", os.Getenv("KISS_COMPRESS"), "")
	pf.BoolVarP(&fs, "no-verify", "s", false, "")
	pf.BoolVarP(&fd, "debug", "d", false, "")
	pf.BoolVarP(&do.Overwrite, "force", "f", false, "")
	pf.BoolVarP(&fn, "no-bar", "n", false, "")
	pf.BoolVarP(&log.NoPrompt, "no-prompt", "y", os.Getenv("KISS_PROMPT") == "1", "")
	// pf.BoolVarP(&bo.NoStripBinaries, "no-strip", "S", os.Getenv("KISS_STRIP") == "0", "")
	pf.BoolVarP(&fT, "no-prebuilt", "T", false, "")
	pf.BoolVarP(&fi, "install", "i", false, "")
	pf.BoolVarP(&fq, "quiet", "q", false, "")

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, buildUsage)
	}

	pf.Parse(args[1:])

	if pf.NArg() == 0 {
		pf.Usage()
		os.Exit(2)
	}

	if fd {
		lo.Debug = true
		bo.Debug = true
	} else {
		// XXX
		defer cleanup.Run(func() error {
			return os.RemoveAll(td)
		})()
	}

	if !fn {
		do.Progress = os.Stderr
	}

	if !fq {
		bo.Output = os.Stdout
	}

	bpp := make([]*king.Package, 0, pf.NArg())

	for _, n := range pf.Args() {
		p, err := king.NewPackage(c, &king.PackageOptions{
			Name: n,
			From: king.All,
		})

		if err != nil {
			return err
		}

		bpp = append(bpp, p)
	}

	epp, dpp, tpp, err := resolveDependencies(c, bpp, fT)

	if err != nil {
		return err
	}

	if len(dpp) > 0 || len(tpp) > 0 {
		// TODO tree?
		w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0) // TODO doc

		fmt.Fprint(w, "<package>\t<type>\t<action>\n")

		for _, t := range tpp {
			fmt.Fprint(w, t.Name+"\t", "pre-built dependency\t", "install\n")
		}

		for _, p := range dpp {
			// TODO print make dependency
			fmt.Fprint(w, p.Name+"\t", "dependency\t", "build && install\n")
		}

		for _, p := range epp {
			fmt.Fprint(w, p.Name+"\t", "candidate\t", "build")

			if fi {
				fmt.Fprint(w, "&& install")
			}

			fmt.Fprint(w, "\n")
		}

		w.Flush()
		log.Prompt("proceed to build?")
	}

	for _, p := range append(dpp, epp...) {
		if err := downloadSources(p, do, fs, fn); err != nil {
			return err
		}
	}

	for _, t := range tpp {
		log.Runningf("installing pre-built dependency %s", t.Name)

		// TODO forcefully install
		// https://github.com/kiss-community/kiss/blob/edfb25aa2da44076dcb35b19f8e6cfddd5a66dfa/kiss#L659
		if _, err := t.Install(lo); err != nil {
			return err
		}
	}

	for _, p := range dpp {
		log.Runningf("building dependency %s", p.Name)

		t, err := p.Build(bo)

		if err != nil {
			return err
		}

		log.Runningf("installing dependency %s", t.Name)

		if _, err := t.Install(lo); err != nil {
			return err
		}
	}

	for _, p := range epp {
		log.Runningf("building %s", p.Name)

		t, err := p.Build(bo)

		if err != nil {
			return err
		}

		if !fi {
			continue
		}

		log.Runningf("installing %s", p.Name)

		if _, err := t.Install(lo); err != nil {
			return err
		}
	}

	return nil
}

// TODO return mpp(make dependencies)
func resolveDependencies(c *king.Config, bpp []*king.Package, fT bool) (epp, dpp []*king.Package,
	tpp []*king.Tarball, err error) {
	mpp := make(map[string]bool, len(bpp))
	epp = make([]*king.Package, 0, len(bpp))

	log.Running("resolving dependencies")

	for _, p := range bpp {
		dd, err := p.RecursiveDependencies()

		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil, err
		}

		for _, d := range dd {
			if mpp[d.Name] {
				continue
			}

			_, err := king.NewPackage(c, &king.PackageOptions{
				Name: d.Name,
				From: king.Database,
			})

			if errors.Is(err, king.ErrPackageNameNotFound) {
				//
			} else if err != nil {
				return nil, nil, nil, err
			} else {
				continue
			}

			p, err := king.NewPackage(c, &king.PackageOptions{
				Name: d.Name,
				From: king.Repository,
			})

			if err != nil {
				return nil, nil, nil, err
			}

			t, err := p.Tarball()

			if fT || errors.Is(err, king.ErrTarballNotFound) {
				dpp = append(dpp, p)
			} else if err != nil {
				return nil, nil, nil, err
			} else {
				tpp = append(tpp, t)
			}

			mpp[p.Name] = true
		}

		if !mpp[p.Name] {
			mpp[p.Name] = true
			epp = append(epp, p)
		}
	}

	return epp, dpp, tpp, nil
}

func downloadSources(p *king.Package, do *king.DownloadOptions, fs, fn bool) error {
	ss, err := p.Sources()

	// TODO inform if sources file not exist
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	dd := make([]king.Downloader, 0, len(ss))

	for _, s := range ss {
		if d, ok := s.(king.Downloader); ok {
			dd = append(dd, d)
		}
	}

	for _, d := range dd {
		if fn {
			// TODO add package name to prefix
			// >> downloading libX11-1.7.0.tar.bz2
			log.Runningf("downloading %s", d)
		}

		// TODO inform if already downloaded
		if err := d.Download(do); err != nil {
			return err
		}
	}

	if fs {
		return nil
	}

	vv := make([]king.Verifier, 0, len(ss))

	for _, s := range ss {
		if v, ok := s.(king.Verifier); ok {
			vv = append(vv, v)
		}
	}

	if len(vv) == 0 {
		return nil
	}

	for _, v := range vv {
		// TODO add package name to prefix
		// >> verifying libX11-1.7.0.tar.bz2
		log.Runningf("verifying %s", v)

		if err := v.Verify(); err != nil {
			return err
		}
	}

	return nil
}
