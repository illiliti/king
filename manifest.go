package king

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/illiliti/king/internal/file"
)

func saveManifest(pp []string, d string) error {
	f, err := os.Create(d)

	if err != nil {
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

func readManifest(s string) ([]string, error) {
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

func generateManifest(d string) ([]string, error) {
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

func installFile(s, d string, m os.FileMode) error {
	switch {
	case m.IsDir():
		if err := os.MkdirAll(d, 0777); err != nil {
			return err
		}

		return os.Chmod(d, m)
	case m.IsRegular():
		return file.CopyRegular(s, d)
	case m&os.ModeSymlink != 0:
		return file.CopySymlink(s, d)
	}

	return nil
}

func removeFile(s string, m os.FileMode) error {
	if m.IsDir() {
		dd, err := file.ReadDirNames(s)

		if err != nil {
			return err
		}

		if len(dd) > 0 {
			return nil
		}
	}

	return os.Remove(s)
}
