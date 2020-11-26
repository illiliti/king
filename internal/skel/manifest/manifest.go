package manifest

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Generate(d string) ([]string, error) {
	var pp []string

	err := filepath.Walk(d, func(p string, st os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		p = strings.TrimPrefix(p, d)

		if p == "" {
			return nil
		}

		if st.IsDir() {
			p += "/"
		}

		pp = append(pp, p)
		return nil
	})

	sort.Sort(sort.Reverse(sort.StringSlice(pp)))
	return pp, err
}

func Replace(s, o, n string) error {
	st, err := os.Stat(s)

	if err != nil {
		return err
	}

	f, err := os.OpenFile(s, os.O_RDWR, st.Mode())

	if err != nil {
		return err
	}

	var pp []string

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		p := sc.Text()

		if p == o {
			p = n
		}

		pp = append(pp, p)
	}

	if err := sc.Err(); err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(pp)))

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	for _, p := range pp {
		fmt.Fprintln(w, p)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return f.Close()
}
