package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
)

func swap(c *king.Config, args []string) error {
	var ft string

	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	pf.StringVarP(&ft, "target", "t", "", "")

	pf.SetInterspersed(true)

	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, swapUsage)
	}

	pf.Parse(args[1:])

	aa, err := parseAlternatives(c, ft, pf.Args())

	if err != nil {
		return err
	}

	for _, a := range aa {
		p, err := king.NewPackage(c, &king.PackageOptions{
			Path: a.Path,
		})

		if err != nil {
			return err
		}

		log.Runningf("swapping %s from %s to %s", a.Path, p.Name, a.Name)

		if _, err := a.Swap(); err != nil {
			return err
		}
	}

	return nil
}

func parseAlternatives(c *king.Config, n string, args []string) ([]*king.Alternative, error) {
	if len(args) > 0 && args[0] != "-" {
		aa := make([]*king.Alternative, 0, len(args))

		// TODO resolve symlinks
		for _, p := range args {
			a, err := king.NewAlternative(c, &king.AlternativeOptions{
				Name: n,
				Path: p,
			})

			if err != nil {
				return nil, err
			}

			aa = append(aa, a)
		}

		return aa, nil
	}

	st, err := os.Stdin.Stat()

	if err != nil {
		return nil, err
	}

	if st.Mode()&os.ModeCharDevice != 0 {
		return nil, errors.New("input is empty")
	}

	var aa []*king.Alternative

	sc := bufio.NewScanner(os.Stdin)

	for sc.Scan() {
		fi := strings.Fields(sc.Text())

		if len(fi) < 2 || fi[0][0] == '#' {
			continue
		}

		a, err := king.NewAlternative(c, &king.AlternativeOptions{
			Name: fi[0],
			Path: fi[1],
		})

		if err != nil {
			return nil, err
		}

		aa = append(aa, a)
	}

	return aa, sc.Err()
}
