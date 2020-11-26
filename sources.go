package king

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

type Source struct {
	Protocol  interface{}
	CustomDir string // TODO rename ?
}

type Git struct {
	URL     string
	RefSpec config.RefSpec

	pkg *Package
}

type HTTP struct {
	URL          string
	Path         string
	HasNoExtract bool

	pkg *Package
}

type File struct {
	Path  string
	IsDir bool // TODO remove ?

	pkg *Package
}

func (p *Package) Sources() ([]*Source, error) {
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

			s := new(Source)

			if len(fi) == 2 {
				s.CustomDir = fi[1]
			}

			switch {
			case strings.HasPrefix(fi[0], "git+"):
				s.Protocol, err = newGit(p, strings.TrimPrefix(fi[0], "git+"))
			case strings.Contains(fi[0], "://"):
				s.Protocol, err = newHTTP(p, fi[0], s.CustomDir)
			default:
				s.Protocol, err = newFile(p, fi[0])
			}

			if err != nil {
				return err
			}

			p.sources = append(p.sources, s)
		}

		return sc.Err()
	})

	return p.sources, err
}

func newGit(p *Package, s string) (*Git, error) {
	i := strings.LastIndexAny(s, "#@")

	if i < 0 {
		return &Git{
			URL:     s,
			RefSpec: config.RefSpec("refs/heads/master:refs/remotes/origin/master"),
			pkg:     p,
		}, nil
	}

	switch s[i:][0] {
	case '#':
		if !plumbing.IsHash(s[i+1:]) {
			return nil, fmt.Errorf("invalid hash: %s", s[i+1:])
		}

		return &Git{
			URL:     s[:i],
			RefSpec: config.RefSpec(s[i+1:] + ":refs/remotes/origin/master"),
			pkg:     p,
		}, nil
	case '@':
		return &Git{
			URL:     s[:i],
			RefSpec: config.RefSpec("refs/heads/" + s[i+1:] + ":refs/remotes/origin/" + s[i+1:]),
			pkg:     p,
		}, nil
	}

	panic("unreachable")
}

func newHTTP(p *Package, s, d string) (*HTTP, error) {
	u, err := url.Parse(s)

	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported protocol: %s", s)
	}

	return &HTTP{
		URL:          s,
		Path:         filepath.Join(p.cfg.SourceDir, p.Name, d, filepath.Base(u.Path)),
		HasNoExtract: u.Query()["no-extract"] != nil,
		pkg:          p,
	}, nil
}

func newFile(p *Package, s string) (*File, error) {
	for _, s := range []string{
		filepath.Join(p.Path, s),
		filepath.Join(p.cfg.RootDir, s),
	} {
		st, err := os.Stat(s)

		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}

		return &File{
			Path:  s,
			IsDir: st.IsDir(),
			pkg:   p,
		}, nil
	}

	return nil, fmt.Errorf("source not found: %s", s)
}
