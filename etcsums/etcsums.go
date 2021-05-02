// Package etcsums provides in-memory structure for etcsums file
package etcsums

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/illiliti/king/internal/hash"
)

// TODO better docs

type Etcsums struct {
	f  *os.File
	ee map[string]bool
}

func Open(p string, o int) (*Etcsums, error) {
	f, err := os.OpenFile(p, o, 0666)

	if err != nil {
		return nil, err
	}

	e := &Etcsums{
		f:  f,
		ee: make(map[string]bool),
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		e.ee[sc.Text()] = true
	}

	return e, sc.Err()
}

func Create(p string) (*Etcsums, error) {
	f, err := os.Create(p)

	if err != nil {
		return nil, err
	}

	return &Etcsums{
		f:  f,
		ee: make(map[string]bool),
	}, nil
}

func (e *Etcsums) Generate(d string) error {
	if e == nil {
		return nil // TODO err
	}

	e.ee = make(map[string]bool, len(e.ee))

	return filepath.WalkDir(d, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !de.Type().IsRegular() {
			return nil
		}

		x, err := hash.Sha256(p)

		if err != nil {
			return err
		}

		e.ee[x] = true
		return nil
	})
}

func (e *Etcsums) Has(x string) bool {
	return e != nil && e.ee[x]
}

func (e *Etcsums) Flush() error {
	if e == nil {
		return nil // TODO err
	}

	if err := e.f.Truncate(0); err != nil {
		return err
	}

	if _, err := e.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	w := bufio.NewWriter(e.f)

	for e := range e.ee {
		w.WriteString(e + "\n")
	}

	return w.Flush()
}

func (e *Etcsums) Close() error {
	if e == nil {
		return nil // TODO err
	}

	return e.f.Close()
}
