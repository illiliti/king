package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func alternative(c *king.Config, args []string) {
	st, err := os.Stdin.Stat()

	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 && st.Mode()&fs.ModeCharDevice != 0 {
		dd, err := os.ReadDir(filepath.Join(c.RootDir, king.ChoicesDir))

		if err != nil {
			log.Fatal(err)
		}

		w := bufio.NewWriter(os.Stdout)
		defer w.Flush()

		for _, de := range dd {
			i := strings.Index(de.Name(), ">")

			if i < 0 {
				continue
			}

			fmt.Fprintln(w, de.Name()[:i], strings.ReplaceAll(de.Name()[i:], ">", "/"))
		}

		return
	}

	for _, a := range func() []*king.Alternative {
		if len(args) == 2 {
			a, err := king.NewAlternativeByNamePath(c, args[0], args[1])

			if err != nil {
				log.Fatal(err)
			}

			return []*king.Alternative{a}
		}

		var aa []*king.Alternative

		sc := bufio.NewScanner(os.Stdin)

		for sc.Scan() {
			fi := strings.Fields(sc.Text())

			if len(fi) < 2 || fi[0][0] == '#' {
				continue
			}

			a, err := king.NewAlternativeByNamePath(c, fi[0], fi[1])

			if err != nil {
				log.Fatal(err)
			}

			aa = append(aa, a)
		}

		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}

		return aa
	}() {
		p, err := king.NewPackageByPath(c, a.Path)

		if err != nil {
			log.Fatal(err)
		}

		log.Runningf("swapping %s from %s to %s", a.Path, p.Name, a.Name)

		if _, err := a.Swap(); err != nil {
			log.Fatal(err)
		}
	}
}
