package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func list(c *king.Config, args []string) {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	for _, n := range func() []string {
		if len(args) > 0 {
			return args
		}

		f, err := os.Open(c.SysDB)

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		dd, err := f.Readdirnames(0)

		if err != nil {
			log.Fatal(err)
		}

		return dd
	}() {
		p, err := king.NewPackageByName(c, king.Sys, n)

		if err != nil {
			log.Fatal(err)
		}

		v, err := p.Version()

		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(w, p.Name, v.Version, v.Release)
	}
}
