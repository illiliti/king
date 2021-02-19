package main

import (
	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func install(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	for _, n := range args {
		t, err := king.NewTarball(c, n)

		if err != nil {
			p, err := king.NewPackageByName(c, king.Any, n)

			if err != nil {
				log.Fatal(err)
			}

			t, err = p.Tarball()

			if err != nil {
				log.Fatal(err)
			}
		}

		log.Runningf("installing %s", t.Name)

		if _, err := t.Install(c.HasForce); err != nil {
			log.Fatal(err)
		}
	}

	log.Infof("installed %s", args)
}
