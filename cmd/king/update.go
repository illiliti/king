package main

import (
	"os"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func update(c *king.Config) {
	upp, err := c.Update()

	if err != nil {
		log.Fatal(err)
	}

	for _, p := range upp {
		dd, err := p.RecursiveDepends()

		if err != nil {
			log.Fatal(err)
		}

		dpp := make([]*king.Package, 0, len(dd))

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

		upp = append(dpp, upp...)
	}

	mpp := make(map[string]bool, len(upp))
	ipp := make([]*king.Package, 0, len(upp))

	for _, p := range upp {
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

	for _, p := range ipp {
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

		t, err := p.Build()

		if err != nil {
			log.Fatal(err)
		}

		if _, err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}
}
