package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/illiliti/king/internal/cleanup"
	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
)

// TODO remove this altogether. i need to think about it

// TODO print how many packages are built/packages in terminal title
// TODO print elapsed time

func update(c *king.Config, td string, args []string) error {
	var (
		fs bool
		fd bool
		fT bool
		fq bool
		fn bool
	)

	bo := new(king.BuildOptions)
	uo := new(king.UpdateOptions)
	lo := new(king.InstallOptions)
	do := new(king.DownloadOptions)

	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	pf.StringVarP(&lo.ExtractDir, "extract-dir", "X", filepath.Join(td, "extract"), "")
	pf.StringVarP(&bo.PackageDir, "package-dir", "P", filepath.Join(td, "pkg"), "")
	// pf.StringVarP(&fO, "output-dir", "O", filepath.Join(cd, "logs"), "")
	pf.StringVarP(&bo.BuildDir, "build-dir", "B", filepath.Join(td, "build"), "")
	pf.StringVarP(&bo.Compression, "compression", "C", "", "")
	pf.StringSliceVarP(&uo.ExcludePackages, "exclude", "x", nil, "")
	pf.BoolVarP(&fs, "no-verify", "s", false, "")
	pf.BoolVarP(&fd, "debug", "d", false, "")
	pf.BoolVarP(&do.Overwrite, "force", "f", false, "")
	pf.BoolVarP(&fn, "no-bar", "n", false, "")
	pf.BoolVarP(&log.NoPrompt, "no-prompt", "y", os.Getenv("KISS_PROMPT") == "1", "")
	pf.BoolVarP(&uo.NoUpdateRepositories, "no-pull", "N", false, "")
	pf.BoolVarP(&uo.ContinueOnError, "no-error", "c", false, "")
	pf.BoolVarP(&fT, "no-prebuilt", "T", false, "")
	pf.BoolVarP(&fq, "quiet", "q", false, "")

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, updateUsage)
	}

	pf.Parse(args[1:])

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

	if !uo.NoUpdateRepositories {
		log.Running("updating repositories")
	}

	upp, err := king.Update(c, uo)

	if err != nil {
		return err
	}

	if len(upp) == 0 {
		log.Info("no one package needing update")
		return nil
	}

	epp, dpp, tpp, err := resolveDependencies(c, upp, fT)

	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0) // TODO doc

	// TODO tree?
	// TODO show old versions

	fmt.Fprint(w, "<package>\t<new version>\t<type>\t<action>\n")

	for _, t := range tpp {
		fmt.Fprint(w, t.Name+"\t", "pre-built dependency\t", "install\n")
	}

	for _, p := range dpp {
		v, err := p.Version()

		if err != nil {
			return err
		}

		// TODO print make dependency
		fmt.Fprint(w, p.Name+"\t", v.String()+"\t", "dependency\t", "build && install\n")
	}

	for _, p := range epp {
		v, err := p.Version()

		if err != nil {
			return err
		}

		fmt.Fprint(w, p.Name+"\t", v.String()+"\t", "candidate\t", "update\n")
	}

	w.Flush()
	log.Prompt("proceed to update?")

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

		log.Runningf("installing %s", p.Name)

		if _, err := t.Install(lo); err != nil {
			return err
		}
	}

	return nil
}
