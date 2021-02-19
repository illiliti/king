package etcsum

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/file"
)

type Etcsum struct {
	f  *os.File
	ee map[string]bool
}

func Open(p string) (*Etcsum, error) {
	f, err := os.OpenFile(p, os.O_RDWR, 0666)

	if err != nil {
		return &Etcsum{}, err
	}

	e := &Etcsum{
		f:  f,
		ee: make(map[string]bool),
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		e.ee[sc.Text()] = true
	}

	return e, sc.Err()
}

func Create(p string) (*Etcsum, error) {
	f, err := os.Create(p)

	return &Etcsum{
		f:  f,
		ee: make(map[string]bool),
	}, err
}

func (e *Etcsum) HasEntry(x string) bool {
	return e.ee[x]
}

func (e *Etcsum) Generate(d string) error {
	e.ee = make(map[string]bool, len(e.ee))

	return filepath.WalkDir(d, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !de.Type().IsRegular() {
			return nil
		}

		x, err := file.Sha256(p)

		if err != nil {
			return err
		}

		e.ee[x] = true
		return nil
	})
}

func (e *Etcsum) Flush() error {
	if err := e.f.Truncate(0); err != nil {
		return err
	}

	if _, err := e.f.Seek(0, 0); err != nil {
		return err
	}

	w := bufio.NewWriter(e.f)

	for e := range e.ee {
		fmt.Fprintln(w, e)
	}

	return w.Flush()
}

func (e *Etcsum) Close() error {
	return e.f.Close()
}
