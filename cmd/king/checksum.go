package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
	"github.com/illiliti/king/checksums"
)

func checksum(c *king.Config, args []string) error {
	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, checksumUsage)
	}

	pf.Parse(args[1:])

	if pf.NArg() == 0 {
		pf.Usage()
		os.Exit(2)
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

		hh := make([]king.Verifier, 0, len(ss))

		for _, s := range ss {
			if h, ok := s.(king.Verifier); ok {
				hh = append(hh, h)
			}
		}

		if len(hh) == 0 {
			log.Infof("no one source of %s needing checksums", p.Name)
		} else {
			if err := generateChecksums(p, hh); err != nil {
				return err
			}
		}
	}

	return nil
}

func generateChecksums(p *king.Package, hh []king.Verifier) error {
	c, err := checksums.Create(filepath.Join(p.Path, "checksums"))

	if err != nil {
		return err
	}

	for _, h := range hh {
		// TODO add package name to prefix
		log.Runningf("computing sha256 of %s", h)

		x, err := h.Sha256()

		if errors.Is(err, king.ErrSha256NotRegular) {
			continue
		}

		if err != nil {
			return err
		}

		c.Insert(x)
	}

	if err := c.Flush(); err != nil {
		return err
	}

	return c.Close()
}
