package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func checksum(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		p, err := c.NewPackage(n, king.Any)

		if err != nil {
			log.Fatal(err)
		}

		ss, err := p.Sources()

		if err != nil {
			continue
		}

		// TODO move to root ?
		f, err := os.Create(filepath.Join(p.Path, "checksums"))

		if err != nil {
			log.Fatal(err)
		}

		w := bufio.NewWriter(f)

		for _, s := range ss {
			if d, ok := s.Protocol.(king.Downloader); ok {
				if err := d.Download(false); err != nil {
					log.Fatal(err)
				}
			}

			c, ok := s.Protocol.(king.Checksumer)

			if !ok {
				continue
			}

			x, err := c.Checksum()

			if err != nil {
				if errors.Is(err, king.ErrIsDir) {
					continue
				}

				log.Fatal(err)
			}

			fmt.Fprintln(w, x)
		}

		if err := w.Flush(); err != nil {
			log.Fatal(err)
		}

		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
