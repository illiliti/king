package main

import (
	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func install(c *king.Context, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		t, err := c.Tarball(n)

		if err != nil {
			log.Fatal(err)
		}

		if err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}
}
