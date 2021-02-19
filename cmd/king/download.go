package main

import (
	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func download(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		p, err := king.NewPackageByName(c, king.Any, n)

		if err != nil {
			log.Fatal(err)
		}

		ss, err := p.Sources()

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

			if err := d.Download(c.HasForce); err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Infof("processed %s", args)
}
