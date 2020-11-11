package main

import (
	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func build(c *king.Context, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	var dpp []*king.Package

	epp := make([]*king.Package, 0, len(args))

	// TODO filter args duplicates ?
	for _, n := range args {
		p, err := c.NewPackage(n, king.AnyDB)

		if err != nil {
			log.Fatal(err)
		}

		pp, err := p.RecursiveDepends()

		if err != nil {
			log.Fatal(err)
		}

		epp = append(epp, p)
		dpp = append(dpp, pp...)
	}

	const (
		unhandled = iota
		candidate
		installed
	)

	mpp := make(map[string]int, len(dpp))

	for i, p := range dpp {
		if mpp[p.Name] == unhandled {
			mpp[p.Name] = candidate

			t, err := c.Tarball(p.Name)

			if err != nil {
				continue
			}

			if err := t.Install(true); err != nil {
				log.Fatal(err)
			}

			mpp[p.Name] = installed
		}

		dpp = append(dpp[:i], dpp[i+1:]...)
	}

	for i, p := range epp {
		if mpp[p.Name] == candidate {
			epp = append(epp[:i], epp[i+1:]...)
		}
	}

	for _, p := range append(dpp, epp...) {
		ss, err := p.Sources()

		if err != nil {
			continue
		}

		for _, s := range ss {
			if err := s.Download(); err != nil {
				log.Fatal(err)
			}
		}

		if err := p.VerifyChecksums(); err != nil {
			log.Fatal(err)
		}
	}

	for _, p := range dpp {
		t, err := p.Build()

		if err != nil {
			log.Fatal(err)
		}

		if err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}

	for _, p := range epp {
		if _, err := p.Build(); err != nil {
			log.Fatal(err)
		}
	}
}
