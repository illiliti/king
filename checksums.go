package king

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TODO generateChecksums ?

func (p *Package) SaveChecksums() error {
	f, err := os.Create(filepath.Join(p.Path, "checksums"))

	if err != nil {
		return err
	}

	ss, err := p.Sources()

	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	for _, s := range ss {
		h, err := s.Checksum()

		if err != nil {
			return err
		}

		if h == "" {
			continue
		}

		fmt.Fprintln(w, h)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return f.Close()
}

func (p *Package) VerifyChecksums() error {
	b, err := ioutil.ReadFile(filepath.Join(p.Path, "checksums"))

	if err != nil {
		return err
	}

	ss, err := p.Sources()

	if err != nil {
		return err
	}

	for _, s := range ss {
		h, err := s.Checksum()

		if err != nil {
			return err
		}

		if h == "" {
			continue
		}

		if !bytes.Contains(b, []byte(h)) {
			return fmt.Errorf("checksum mismatch: %s", h)
		}
	}

	return nil
}
