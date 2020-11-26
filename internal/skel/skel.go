package skel

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func Map(s string) (map[string]bool, error) {
	f, err := os.Open(s)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	ee := make(map[string]bool)
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		ee[sc.Text()] = true
	}

	return ee, sc.Err()
}

func Slice(s string) ([]string, error) {
	f, err := os.Open(s)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	var pp []string

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		pp = append(pp, sc.Text())
	}

	return pp, sc.Err()
}

func Save(pp []string, w io.Writer) error {
	bw := bufio.NewWriter(w)

	for _, p := range pp {
		fmt.Fprintln(bw, p)
	}

	return bw.Flush()
}
