package king

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/fetch"
	"github.com/illiliti/king/internal/file"
)

type Source struct {
	Protocol  interface{}
	CustomDir string
}

type File struct {
	Path  string
	IsDir bool
}

type Git struct {
	URL string
}

type HTTP struct {
	URL          string
	Path         string // TODO refactor ?
	IsCached     bool
	HasNoExtract bool
}

// TODO DownloadSources ?

func (p *Package) Sources() ([]*Source, error) {
	newSource := func(s, d string) (*Source, error) {
		if strings.HasPrefix(s, "git+") {
			return &Source{Protocol: &Git{
				URL: strings.TrimPrefix(s, "git+"),
			}}, nil
		}

		if strings.Contains(s, "://") {
			u, err := url.Parse(s)

			if err != nil {
				return nil, err
			}

			if u.Scheme != "http" && u.Scheme != "https" {
				return nil, fmt.Errorf("unsupported protocol: %s", s)
			}

			cs := filepath.Join(p.context.SourceDir, p.Name, d, filepath.Base(u.Path))
			ns := &Source{Protocol: &HTTP{
				URL:  s,
				Path: cs,
			}}

			if _, ok := u.Query()["no-extract"]; ok {
				ns.Protocol.(*HTTP).HasNoExtract = true
			}

			if _, err := os.Stat(cs); !os.IsNotExist(err) {
				ns.Protocol.(*HTTP).IsCached = true
			}

			return ns, nil
		}

		for _, l := range []string{
			filepath.Join(p.Path, s),
			filepath.Join(p.context.RootDir, s),
		} {
			st, err := os.Stat(l)

			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return nil, err
			}

			return &Source{Protocol: &File{
				Path:  l,
				IsDir: st.IsDir(),
			}}, nil
		}

		return nil, fmt.Errorf("invalid source: %s", s)
	}

	err := p.sourcesOnce.Do(func() error {
		f, err := os.Open(filepath.Join(p.Path, "sources"))

		if err != nil {
			return err
		}

		defer f.Close()

		sc := bufio.NewScanner(f)

		for sc.Scan() {
			fi := strings.Fields(sc.Text())

			if len(fi) == 0 || fi[0][0] == '#' {
				continue
			}

			var d string

			if len(fi) == 2 {
				d = fi[1]
			}

			s, err := newSource(fi[0], d)

			if err != nil {
				return err
			}

			s.CustomDir = d // XXX i hate this
			p.sources = append(p.sources, s)
		}

		return sc.Err()
	})

	return p.sources, err
}

func (s *Source) Checksum() (string, error) {
	switch v := s.Protocol.(type) {
	case *Git:
		return "", nil
	case *HTTP:
		return file.Sha256Sum(v.Path)
	case *File:
		if v.IsDir {
			return "", nil
		}

		return file.Sha256Sum(v.Path)
	}

	panic("unreachable")
}

func (s *Source) Download() error {
	switch v := s.Protocol.(type) {
	case *Git:
		// found git source
	case *File:
		// found local source
	case *HTTP:
		if v.IsCached {
			// found cached source
			return nil
		}

		return fetch.HTTPDownload(v.URL, v.Path)
	}

	return nil
}
