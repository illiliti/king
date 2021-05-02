// Package manifest provides in-memory structure for manifest file
package manifest

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TODO unit tests
// TODO better docs

type SortType uint

const (
	NoSort SortType = iota
	Directories
	Files
)

type Manifest struct {
	f  *os.File
	pp map[string]bool
}

func Open(p string, o int) (*Manifest, error) {
	f, err := os.OpenFile(p, o, 0666)

	if err != nil {
		return nil, err
	}

	m := &Manifest{
		f:  f,
		pp: make(map[string]bool),
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		m.pp[sc.Text()] = true
	}

	return m, sc.Err()
}

func Create(p string) (*Manifest, error) {
	f, err := os.Create(p)

	if err != nil {
		return nil, err
	}

	return &Manifest{
		f:  f,
		pp: make(map[string]bool),
	}, nil
}

func (m *Manifest) Sort(st SortType) []string {
	if m == nil {
		return nil // TODO err
	}

	pp := make([]string, 0, len(m.pp))

	for p := range m.pp {
		pp = append(pp, p)
	}

	switch st {
	case NoSort:
		//
	case Directories:
		sort.Strings(pp)
	case Files:
		sort.Sort(sort.Reverse(sort.StringSlice(pp)))
	}

	return pp
}

func (m *Manifest) Generate(d string) error {
	if m == nil {
		return nil // TODO err
	}

	m.pp = make(map[string]bool, len(m.pp))

	return filepath.WalkDir(d, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		p = strings.TrimPrefix(p, d) // FIXME i believe we should avoid this

		if p == "" {
			return nil
		}

		if de.IsDir() {
			p += "/"
		}

		m.pp[p] = true
		return nil
	})
}

func (m *Manifest) Has(p string) bool {
	return m != nil && m.pp[p]
}

func (m *Manifest) Delete(p string) {
	if m == nil {
		return
	}

	delete(m.pp, p)
}

func (m *Manifest) Insert(p string) {
	if m == nil {
		return
	}

	m.pp[p] = true
}

func (m *Manifest) Replace(o, n string) {
	if m == nil {
		return
	}

	m.Delete(o)
	m.Insert(n)
}

func (m *Manifest) Rehash() error {
	if m == nil {
		return nil // TODO err
	}

	if _, err := m.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	m.pp = make(map[string]bool, len(m.pp))

	sc := bufio.NewScanner(m.f)

	for sc.Scan() {
		m.pp[sc.Text()] = true
	}

	return sc.Err()
}

func (m *Manifest) Flush() error {
	if m == nil {
		return nil // TODO err
	}

	if err := m.f.Truncate(0); err != nil {
		return err
	}

	if _, err := m.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	w := bufio.NewWriter(m.f)

	for _, p := range m.Sort(Files) {
		w.WriteString(p + "\n")
	}

	return w.Flush()
}

func (m *Manifest) Close() error {
	if m == nil {
		return nil // TODO err
	}

	return m.f.Close()
}
