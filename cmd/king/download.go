package main

import (
	"strings"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func download(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		p, err := c.NewPackageByName(king.Any, n)

		if err != nil {
			log.Fatal(err)
		}

		ss, err := p.Sources()

		if err != nil {
			log.Fatal(err)
		}

		log.Runningf("downloading %s", p.Name)

		for _, s := range ss {
			d, ok := s.Protocol.(king.Downloader)

			if !ok {
				continue
			}

			if err := d.Download(c.HasForce); err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Successf("downloaded %s", strings.Join(args, ", "))
}
