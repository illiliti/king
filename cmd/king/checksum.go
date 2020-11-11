package main

import (
	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func checksum(c *king.Context, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		p, err := c.NewPackage(n, king.AnyDB)

		if err != nil {
			log.Fatal(err)
		}

		ss, err := p.Sources()

		if err != nil {
			log.Warning(err)
			continue
		}

		for _, s := range ss {
			if err := s.Download(); err != nil {
				log.Fatal(err)
			}
		}

		if err := p.SaveChecksums(); err != nil {
			log.Fatal(err)
		}
	}
}
