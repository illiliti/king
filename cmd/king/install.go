package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/cleanup"
	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
)

// TODO prompt to confirm
func install(c *king.Config, td string, args []string) error {
	lo := new(king.InstallOptions)

	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	pf.BoolVarP(&lo.Debug, "debug", "d", false, "")
	pf.BoolVarP(&lo.NoCheckDependencies, "force", "f", os.Getenv("KISS_FORCE") == "1", "")
	pf.BoolVarP(&lo.OverwriteEtcFiles, "overwrite-etc", "e", false, "")
	pf.StringVarP(&lo.ExtractDir, "extract-dir", "X", filepath.Join(td, "extract"), "")

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, installUsage)
	}

	pf.Parse(args[1:])

	if pf.NArg() == 0 {
		pf.Usage()
		os.Exit(2)
	}

	// XXX
	if !lo.Debug {
		defer cleanup.Run(func() error {
			return os.RemoveAll(td)
		})()
	}

	for _, n := range pf.Args() {
		t, err := king.NewTarball(c, n)

		if err != nil {
			p, err := king.NewPackage(c, &king.PackageOptions{
				Name: n,
				From: king.All,
			})

			if err != nil {
				return err
			}

			t, err = p.Tarball()

			if err != nil {
				return err
			}
		}

		log.Runningf("installing %s", t.Name)

		if _, err := t.Install(lo); err != nil {
			return err
		}
	}

	return nil
}
