package chksum

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Chksum struct {
	f  *os.File
	cc map[string]bool
}

func Open(p string) (*Chksum, error) {
	f, err := os.OpenFile(p, os.O_RDWR, 0666)

	if err != nil {
		return &Chksum{}, err
	}

	c := &Chksum{
		f:  f,
		cc: make(map[string]bool),
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		// TODO drop backward compatibility for kiss 4.x.x
		c.cc[strings.Fields(sc.Text())[0]] = true
	}

	return c, sc.Err()
}

func Create(p string) (*Chksum, error) {
	f, err := os.Create(p)

	return &Chksum{
		f:  f,
		cc: make(map[string]bool),
	}, err
}

func (c *Chksum) HasEntry(x string) bool {
	return c.cc[x]
}

func (c *Chksum) Insert(x string) {
	if x != "" {
		c.cc[x] = true
	}
}

func (c *Chksum) Flush() error {
	if err := c.f.Truncate(0); err != nil {
		return err
	}

	if _, err := c.f.Seek(0, 0); err != nil {
		return err
	}

	w := bufio.NewWriter(c.f)

	for c := range c.cc {
		fmt.Fprintln(w, c)
	}

	return w.Flush()
}

func (c *Chksum) Close() error {
	return c.f.Close()
}
