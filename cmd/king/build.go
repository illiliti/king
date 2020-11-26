package main

import (
	"os"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func build(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	var dpp []*king.Package

	epp := make([]*king.Package, 0, len(args))

	// TODO filter args duplicates ?
	for _, n := range args {
		p, err := c.NewPackage(n, king.Any)

		if err != nil {
			log.Fatal(err)
		}

		dd, err := p.RecursiveDepends()

		if err != nil {
			log.Fatal(err)
		}

		for _, d := range dd {
			if _, err := c.NewPackage(d.Name, king.Sys); err == nil {
				continue
			}

			p, err := c.NewPackage(d.Name, king.User)

			if err != nil {
				log.Fatal(err)
			}

			dpp = append(dpp, p)
		}

		epp = append(epp, p)
	}

	mpp := make(map[string]bool, len(dpp))
	ipp := make([]*king.Package, 0, len(dpp))

	// TODO redo
	for _, p := range dpp {
		if mpp[p.Name] {
			continue
		}

		mpp[p.Name] = true

		t, err := c.Tarball(p.Name)

		if err != nil {
			ipp = append(ipp, p) // TODO redo
			continue
		}

		if _, err := t.Install(true); err != nil {
			log.Fatal(err)
		}
	}

	bpp := make([]*king.Package, 0, len(epp))

	for _, p := range epp {
		if !mpp[p.Name] {
			bpp = append(bpp, p)
		}
	}

	for _, p := range append(ipp, bpp...) {
		ss, err := p.Sources()

		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}

		for _, s := range ss {
			if d, ok := s.Protocol.(king.Downloader); ok {
				if err := d.Download(false); err != nil {
					log.Fatal(err)
				}
			}

			if v, ok := s.Protocol.(king.Checksumer); ok {
				if err := v.Verify(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	for _, p := range ipp {
		t, err := p.Build()

		if err != nil {
			log.Fatal(err)
		}

		if _, err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}

	for _, p := range bpp {
		if _, err := p.Build(); err != nil {
			log.Fatal(err)
		}
	}
}
