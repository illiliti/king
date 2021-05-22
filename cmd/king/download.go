package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
)

// TODO download dependencies too

func download(c *king.Config, args []string) error {
	var fn bool

	do := new(king.DownloadOptions)

	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	pf.BoolVarP(&do.Overwrite, "force", "f", false, "")
	pf.BoolVarP(&fn, "no-bar", "n", false, "")

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, downloadUsage)
	}

	pf.Parse(args[1:])

	if pf.NArg() == 0 {
		pf.Usage()
		os.Exit(2)
	}

	if !fn {
		do.Progress = os.Stderr
	}

	for _, n := range pf.Args() {
		p, err := king.NewPackage(c, &king.PackageOptions{
			Name: n,
			From: king.All,
		})

		if err != nil {
			return err
		}

		ss, err := p.Sources()

		// TODO skip ?
		if err != nil {
			return err
		}

		dd := make([]king.Downloader, 0, len(ss))

		for _, s := range ss {
			if d, ok := s.(king.Downloader); ok {
				dd = append(dd, d)
			}
		}

		if len(dd) == 0 {
			log.Infof("no one source of %s needing download", p.Name)
		} else {
			for _, d := range dd {
				if fn {
					log.Runningf("downloading %s", d)
				}

				err := d.Download(do)

				if errors.Is(err, king.NoErrDownloadAlreadyDownloaded) {
					log.Info(err)
				} else if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
