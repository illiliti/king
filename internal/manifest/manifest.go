package manifest

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Manifest struct {
	f  *os.File
	pp []string
}

func Open(p string) (*Manifest, error) {
	f, err := os.OpenFile(p, os.O_RDWR, 0666)

	if err != nil {
		return &Manifest{}, err
	}

	m := &Manifest{
		f: f,
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		m.pp = append(m.pp, sc.Text())
	}

	return m, sc.Err()
}

func Create(p string) (*Manifest, error) {
	f, err := os.Create(p)

	return &Manifest{
		f: f,
	}, err
}

func (m *Manifest) Install() []string {
	sort.Strings(m.pp)
	return m.pp
}

func (m *Manifest) Remove() []string {
	sort.Sort(sort.Reverse(sort.StringSlice(m.pp)))
	return m.pp
}

func (m *Manifest) HasEntry(e string) bool {
	for _, p := range m.pp {
		if p == e {
			return true
		}
	}

	return false
}

func (m *Manifest) Generate(d string) error {
	m.pp = make([]string, 0, len(m.pp))

	return filepath.WalkDir(d, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		p = strings.TrimPrefix(p, d)

		if p == "" {
			return nil
		}

		if de.IsDir() {
			p += "/"
		}

		m.pp = append(m.pp, p)
		return nil
	})
}

func (m *Manifest) Replace(o, n string) {
	for i, p := range m.pp {
		if p == o {
			m.pp[i] = n
		}
	}
}

func (m *Manifest) Insert(p string) {
	if !m.HasEntry(p) {
		m.pp = append(m.pp, p)
	}
}

func (m *Manifest) Delete(p string) {
	for i, o := range m.pp {
		if o == p {
			m.pp[i] = m.pp[len(m.pp)-1]
			m.pp = m.pp[:len(m.pp)-1]
		}
	}
}

func (m *Manifest) Flush() error {
	if err := m.f.Truncate(0); err != nil {
		return err
	}

	if _, err := m.f.Seek(0, 0); err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(m.pp)))

	w := bufio.NewWriter(m.f)

	for _, p := range m.pp {
		fmt.Fprintln(w, p)
	}

	return w.Flush()
}

func (m *Manifest) Close() error {
	return m.f.Close()
}
