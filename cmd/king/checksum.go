package main

import (
	"path/filepath"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/chksum"
	"github.com/illiliti/king/internal/log"
)

func checksum(c *king.Config, args []string) {
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

		hh := make([]king.Checksum, 0, len(ss))

		log.Runningf("preparing %s", p.Name)

		for _, s := range ss {
			h, ok := s.(king.Checksum)

			if !ok {
				log.Infof("skipping %s", s)
			} else {
				hh = append(hh, h)
			}
		}

		if len(hh) == 0 {
			continue
		}

		c, err := chksum.Create(filepath.Join(p.Path, "checksums"))

		if err != nil {
			log.Fatal(err)
		}

		for _, h := range hh {
			log.Runningf("generating %s", h)

			x, err := h.Sha256()

			if err != nil {
				log.Fatal(err)
			}

			c.Insert(x)
		}

		if err := c.Flush(); err != nil {
			log.Fatal(err)
		}

		if err := c.Close(); err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("processed %s", args)
}
