package main

import (
	"errors"
	"io/fs"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func build(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	var (
		dpp []*king.Package
		tpp []*king.Tarball
	)

	mpp := make(map[string]bool)
	epp := make([]*king.Package, 0, len(args))

	log.Running("resolving dependencies")

	for _, n := range args {
		p, err := king.NewPackageByName(c, king.Any, n)

		if err != nil {
			log.Fatal(err)
		}

		dd, err := p.RecursiveDepends()

		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Fatal(err)
		}

		for _, d := range dd {
			if mpp[d.Name] {
				continue
			}

			if _, err := king.NewPackageByName(c, king.Sys, d.Name); err == nil {
				continue
			}

			p, err := king.NewPackageByName(c, king.Usr, d.Name)

			if err != nil {
				log.Fatal(err)
			}

			t, err := p.Tarball()

			if err != nil {
				dpp = append(dpp, p)
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

	app := append(dpp, epp...)

	log.Askf("proceed to build? %s", app)

	for _, p := range app {
		ss, err := p.Sources()

		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			log.Fatal(err)
		}

		log.Runningf("preparing %s", p.Name)

		for _, s := range ss {
			d, ok := s.(king.Downloader)

			if !ok {
				continue
			}

			log.Runningf("downloading %s", d)

			if err := d.Download(false); err != nil {
				log.Fatal(err)
			}
		}

		for _, s := range ss {
			c, ok := s.(king.Checksum)

			if !ok {
				continue
			}

			log.Runningf("verifying %s", c)

			if err := c.Verify(); err != nil {
				log.Fatal(err)
			}
		}
	}

	for _, t := range tpp {
		log.Runningf("installing pre-built dependency %s", t.Name)

		if _, err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}

	for _, p := range dpp {
		log.Runningf("building dependency %s", p.Name)

		t, err := p.Build()

		if err != nil {
			log.Fatal(err)
		}

		log.Runningf("installing dependency %s", t.Name)

		if _, err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}

	for _, p := range epp {
		log.Runningf("building %s", p.Name)

		if _, err := p.Build(); err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("processed %s", args)
}
