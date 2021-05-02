package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
)

func remove(c *king.Config, args []string) error {
	var fr bool

	ro := new(king.RemoveOptions)

	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	pf.BoolVarP(&ro.NoCheckReverseDependencies, "force", "f", os.Getenv("KISS_FORCE") == "1", "")
	pf.BoolVarP(&ro.NoSwapAlternatives, "no-swap", "a", false, "")
	pf.BoolVarP(&ro.RemoveEtcFiles, "remove-etc", "e", false, "")
	pf.BoolVarP(&fr, "recursive", "r", false, "")

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, removeUsage)
	}

	pf.Parse(args[1:])

	if pf.NArg() == 0 {
		pf.Usage()
		os.Exit(2)
	}

	for _, n := range pf.Args() {
		p, err := king.NewPackage(c, &king.PackageOptions{
			Name: n,
			From: king.Database,
		})

		if err != nil {
			return err
		}

		if fr {
			dd, err := p.RecursiveDependencies()

			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}

			for _, d := range dd {
				p, err := king.NewPackage(c, &king.PackageOptions{
					Name: d.Name,
					From: king.Database,
				})

				if err != nil {
					return err
				}

				log.Runningf("removing dependency %s", p.Name)

				if err := p.Remove(ro); err != nil {
					if errors.Is(err, king.ErrRemoveUnresolvedDependencies) {
						log.Info(err)
					} else {
						return err
					}
				}
			}
		}

		log.Runningf("removing %s", p.Name)

		if err := p.Remove(ro); err != nil {
			return err
		}
	}

	return nil
}
