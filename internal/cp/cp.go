package cp

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TODO unit tests

func CopyLink(s, d string) error {
	l, err := os.Readlink(s)

	if err != nil {
		return err
	}

	if err := os.Remove(d); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(d), 0777); err != nil {
		return err
	}

	return os.Symlink(l, d)
}

func CopyFile(s, d string) error {
	if st, err := os.Stat(d); err == nil && st.IsDir() {
		d = filepath.Join(d, filepath.Base(s))
	}

	st, err := os.Stat(s)

	if err != nil {
		return err
	}

	if err := os.Remove(d); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	sf, err := os.Open(s)

	if err != nil {
		return err
	}

	defer sf.Close()

	sd, err := os.OpenFile(d, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0)

	if err != nil {
		return err
	}

	if err := sd.Chmod(st.Mode()); err != nil {
		return err
	}

	if _, err := io.Copy(sd, sf); err != nil {
		return err
	}

	return sd.Close()
}

func CopyDir(s, d string) error {
	return filepath.WalkDir(s, func(p string, de os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		dp := filepath.Join(d, strings.TrimPrefix(p, s)) // FIXME i believe we should avoid this

		switch {
		case de.IsDir():
			if err := os.MkdirAll(dp, 0777); err != nil {
				return err
			}

			st, err := de.Info()

			if err != nil {
				return err
			}

			err = os.Chmod(dp, st.Mode())
		case de.Type().IsRegular():
			err = CopyFile(p, dp)
		case de.Type()&os.ModeSymlink != 0:
			err = CopyLink(p, dp)
		}

		return err
	})
}
