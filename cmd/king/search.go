package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func search(c *king.Config, args []string) {
	if len(args) == 0 {
		log.Fatal("not enough arguments")
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	for _, n := range args {
		for _, db := range append(c.UserDB, c.SysDB) {
			mm, err := filepath.Glob(filepath.Join(db, n))

			if err != nil {
				log.Fatal(err)
			}

			for _, m := range mm {
				fmt.Fprintln(w, m)
			}
		}
	}
}
