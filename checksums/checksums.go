// Package checksums provides in-memory structure for checksums file
package checksums

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// TODO better docs

type Checksums struct {
	f  *os.File
	cc map[string]bool
}

func Open(p string, o int) (*Checksums, error) {
	f, err := os.OpenFile(p, o, 0666)

	if err != nil {
		return nil, err
	}

	c := &Checksums{
		f:  f,
		cc: make(map[string]bool),
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		// TODO drop backward compatibility for kiss 4.x.x
		// TODO ban SKIP
		c.cc[strings.Fields(sc.Text())[0]] = true
	}

	return c, sc.Err()
}

func Create(p string) (*Checksums, error) {
	f, err := os.Create(p)

	if err != nil {
		return nil, err
	}

	return &Checksums{
		f:  f,
		cc: make(map[string]bool),
	}, nil
}

func (c *Checksums) Has(x string) bool {
	return c != nil && c.cc[x]
}

func (c *Checksums) Insert(x string) {
	if c == nil || x == "" {
		return
	}

	c.cc[x] = true
}

func (c *Checksums) Flush() error {
	if c == nil {
		return nil // TODO err
	}

	if err := c.f.Truncate(0); err != nil {
		return err
	}

	if _, err := c.f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// number of hashes + (length of sha256 in hex + newline)
	w := bufio.NewWriterSize(c.f, len(c.cc)*(64+1))

	for c := range c.cc {
		w.WriteString(c + "\n")
	}

	return w.Flush()
}

func (c *Checksums) Close() error {
	if c == nil {
		return nil
	}

	return c.f.Close()
}
