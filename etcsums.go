package king

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/file"
)

func saveEtcsums(ee []string, d string) error {
	f, err := os.Create(d)

	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	for _, e := range ee {
		fmt.Fprintln(w, e)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return f.Close()
}

func readEtcsums(s string) (map[string]bool, error) {
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

func generateEtcsums(d string) ([]string, error) {
	var ee []string

	err := filepath.Walk(filepath.Join(d, "etc"), func(p string,
		st os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !st.Mode().IsRegular() {
			return nil
		}

		h, err := file.Sha256Sum(p)

		if err != nil {
			return err
		}

		ee = append(ee, h)
		return nil
	})

	return ee, err
}
