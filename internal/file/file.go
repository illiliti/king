package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
)

func ReadDirNames(d string) ([]string, error) {
	f, err := os.Open(d)

	if err != nil {
		return nil, err
	}

	defer f.Close()
	return f.Readdirnames(0)
}

func Sha256Sum(p string) (string, error) {
	f, err := os.Open(p)

	if err != nil {
		return "", err
	}

	defer f.Close()

	c := sha256.New()

	if _, err := io.Copy(c, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(c.Sum(nil)), nil
}

// TODO ?
// func Replace(s, o, n string) error {
// 	st, err := os.Stat(s)

// 	if err != nil {
// 		return err
// 	}

// 	f, err := os.OpenFile(s, os.O_RDWR, st.Mode())

// 	if err != nil {
// 		return err
// 	}

// 	var pp []string

// 	sc := bufio.NewScanner(f)

// 	for sc.Scan() {
// 		p := sc.Text()

// 		if p == o {
// 			p = n
// 		}

// 		pp = append(pp, p)
// 	}

// 	if err := sc.Err(); err != nil {
// 		return err
// 	}

// 	sort.Sort(sort.Reverse(sort.StringSlice(pp)))

// 	w := bufio.NewWriter(f)

// 	for _, p := range pp {
// 		fmt.Fprintln(f, p)
// 	}

// 	if err := w.Flush(); err != nil {
// 		return err
// 	}

// 	return f.Close()
// }

func CopySymlink(s, d string) error {
	l, err := os.Readlink(s)

	if err != nil {
		return err
	}

	if err := os.Remove(d); err != nil && !os.IsNotExist(err) {
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

	if err := os.Remove(d); err != nil && !os.IsNotExist(err) {
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
	return filepath.Walk(s, func(p string, st os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dp := filepath.Join(d, strings.TrimPrefix(p, s))

		switch m := st.Mode(); {
		case m.IsDir():
			if err := os.MkdirAll(dp, 0777); err != nil {
				return err
			}

			return os.Chmod(dp, m)
		case m.IsRegular():
			return CopyFile(p, dp)
		case m&os.ModeSymlink != 0:
			return CopySymlink(p, dp)
		}

		return nil
	})
}

func CreateArchive(s, d string) error {
	if err := os.MkdirAll(filepath.Dir(d), 0777); err != nil {
		return err
	}

	w, err := os.Getwd()

	if err != nil {
		return err
	}

	// XXX why ?
	if err := os.Chdir(s); err != nil {
		return err
	}

	defer os.Chdir(w)
	return archiver.Archive([]string{"."}, d)
}

func ExtractArchive(s, d string, c int) error {
	f, err := archiver.ByExtension(s)

	if err != nil {
		return err
	}

	switch v := f.(type) {
	case *archiver.Zip:
		v.StripComponents = c
	case *archiver.Rar:
		v.StripComponents = c
	case *archiver.Tar:
		v.StripComponents = c
	case *archiver.TarBrotli:
		v.Tar.StripComponents = c
	case *archiver.TarBz2:
		v.Tar.StripComponents = c
	case *archiver.TarGz:
		v.Tar.StripComponents = c
	case *archiver.TarLz4:
		v.Tar.StripComponents = c
	case *archiver.TarSz:
		v.Tar.StripComponents = c
	case *archiver.TarXz:
		v.Tar.StripComponents = c
	case *archiver.TarZstd:
		v.Tar.StripComponents = c
	default:
		return fmt.Errorf("unsupported format: %s", f)
	}

	u, ok := f.(archiver.Unarchiver)

	if !ok {
		return fmt.Errorf("archive can't be extracted: %s", f)
	}

	return u.Unarchive(s, d)
}
