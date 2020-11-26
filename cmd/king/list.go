package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/log"
)

func list(c *king.Config, args []string) {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	for _, n := range func() []string {
		if len(args) > 0 {
			return args
		}

		nn, err := file.ReadDirNames(c.SysDB)

		if err != nil {
			log.Fatal(err)
		}

		return nn
	}() {
		p, err := c.NewPackage(n, king.Sys)

		if err != nil {
			log.Fatal(err)
		}

		v, err := p.Version()

		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(w, p.Name, v.Current, v.Release)
	}
}
