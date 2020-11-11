package main

import (
	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func remove(c *king.Context, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		p, err := c.NewPackage(n, king.SysDB)

		if err != nil {
			log.Fatal(err)
		}

		if err := p.Remove(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}
}
