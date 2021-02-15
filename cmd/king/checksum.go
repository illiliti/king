package main

import (
	"errors"
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
		p, err := c.NewPackageByName(king.Any, n)

		if err != nil {
			log.Fatal(err)
		}

		ss, err := p.Sources()

		if err != nil {
			log.Fatal(err)
		}

		hh := make([]king.Checksum, 0, len(ss))

		for _, s := range ss {
			if h, ok := s.Protocol.(king.Checksum); ok {
				hh = append(hh, h)
			}
		}

		if len(hh) == 0 {
			// log.Infof("... %s", p.Name)
			continue
		}

		// log.Runningf("generating checksums %s", p.Name)

		c, err := chksum.Create(filepath.Join(p.Path, "checksums"))

		if err != nil {
			log.Fatal(err)
		}

		for _, h := range hh {
			x, err := h.Sha256()

			if err != nil {
				if errors.Is(err, king.ErrIsDir) {
					continue
				}

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

	// log.Successf("generated checksums %s", strings.Join(args, ", "))
}
