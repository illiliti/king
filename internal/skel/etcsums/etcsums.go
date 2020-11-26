package etcsums

import (
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/file"
)

func Generate(d string) ([]string, error) {
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
