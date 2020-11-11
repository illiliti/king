package main

import (
	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func download(c *king.Context, args []string) {
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
			log.Fatal(err)
		}

		for _, s := range ss {
			if err := s.Download(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
