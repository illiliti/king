package main

import (
	"strings"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func remove(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		p, err := c.NewPackageByName(king.Sys, n)

		if err != nil {
			log.Fatal(err)
		}

		log.Runningf("removing %s", p.Name)

		if err := p.Remove(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}

	log.Successf("removed %s", strings.Join(args, ", "))
}
