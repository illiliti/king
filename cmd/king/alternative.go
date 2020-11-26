package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

func alternative(c *king.Config, args []string) {
	st, err := os.Stdin.Stat()

	if err != nil {
		log.Fatal(err)
	}

	swap := func(np []string) {
		if len(np) != 2 {
			log.Fatal("not enough arguments")
		}

		a, err := c.NewAlternative(np[0], np[1])

		if err != nil {
			log.Fatal(err)
		}

		if _, err := a.Swap(); err != nil {
			log.Fatal(err)
		}
	}

	switch {
	case st.Mode()&os.ModeCharDevice == 0:
		sc := bufio.NewScanner(os.Stdin)

		for sc.Scan() {
			swap(strings.Fields(sc.Text()))
		}

		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}
	case len(args) > 0:
		swap(args)
	default:
		aa, err := c.Alternatives()

		if err != nil {
			log.Fatal(err)
		}

		w := bufio.NewWriter(os.Stdout)
		defer w.Flush()

		for _, a := range aa {
			fmt.Fprintln(w, a.Name, a.Path)
		}
	}
}
