package main

import (
	"errors"
	"io/fs"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func update(c *king.Config) {
	log.Running("updating repositories")

	upp, err := c.Update()

	if err != nil {
		log.Fatal(err)
	}

	var (
		dpp []*king.Package
		tpp []*king.Tarball
	)

	mpp := make(map[string]bool)
	epp := make([]*king.Package, 0, len(upp))

	log.Running("resolving dependencies")

	for _, p := range upp {
		dd, err := p.RecursiveDepends()

		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Fatal(err)
		}

		for _, d := range dd {
			if mpp[d.Name] {
				continue
			}

			if _, err := c.NewPackageByName(king.Sys, d.Name); err == nil {
				continue
			}

			p, err := c.NewPackageByName(king.Usr, d.Name)

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

	// log.Infof("updating %s", strings.Join(ann, ", "))
	// log.Ask("proceed to update?")
	// log.Ask("proceed to update? [enter/ctrl+c]")
	// log.Ask("ready to update %s, press enter to confirm or ctrl+c to abort", strings.Join(..., ", "))

	for _, p := range append(dpp, epp...) {
		ss, err := p.Sources()

		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Fatal(err)
		}

		log.Runningf("downloading %s", p.Name)

		for _, s := range ss {
			d, ok := s.Protocol.(king.Downloader)

			if !ok {
				continue
			}

			if err := d.Download(false); err != nil {
				log.Fatal(err)
			}
		}

		log.Runningf("verifying %s", p.Name)

		for _, s := range ss {
			c, ok := s.Protocol.(king.Checksum)

			if !ok {
				continue
			}

			if err := c.Verify(); err != nil {
				log.Fatal(err)
			}
		}
	}

	// log.Successf("downloaded %s", strings.Join(..., ", "))

	for _, p := range dpp {
		log.Runningf("building dependency %s", p.Name)

		t, err := p.Build()

		if err != nil {
			log.Fatal(err)
		}

		tpp = append(tpp, t)
	}

	// log.Successf("built %s", strings.Join(..., ", "))

	for _, t := range tpp {
		log.Runningf("installing dependency %s", t.Name)

		if _, err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}

	// log.Successf("installed %s", strings.Join(..., ", "))

	for _, p := range epp {
		log.Runningf("building %s", p.Name)

		if _, err := p.Build(); err != nil {
			log.Fatal(err)
		}
	}

	// log.Successf("built %s", strings.Join(..., ", "))

	for _, p := range epp {
		log.Runningf("installing %s", p.Name)

		if _, err := p.Build(); err != nil {
			log.Fatal(err)
		}
	}

	// log.Successf("installed %s", strings.Join(..., ", "))
}
