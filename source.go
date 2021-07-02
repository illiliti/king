package king

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// TODO better docs

var (
	ErrSourceGitSchemeInvalid  = errors.New("target scheme cannot be file")
	ErrSourceGitHashInvalid    = errors.New("target contains invalid commit hash")
	ErrSourceHTTPSchemeInvalid = errors.New("target scheme must be a http or https")
	ErrSourceFileNotFound      = errors.New("target not found in relative or absolute location")
)

// Source represents package source.
// See https://k1sslinux.org/package-system#4.0
type Source interface {
	String() string

	Extractor
}

// HTTP represents source that can be downloaded.
type HTTP struct {
	URL string

	ed string

	cs string
	ne bool

	pkg *Package
}

// File represents absolute or relative (to the path of package) file source.
type File struct {
	Path string

	ed string

	pkg *Package
}

// Git represents git source.
type Git struct {
	URL string

	ed string

	rs []config.RefSpec
}

// Sources parses package sources. Lines that starts with '#' will be ignored.
func (p *Package) Sources() ([]Source, error) {
	f, err := os.Open(filepath.Join(p.Path, "sources"))

	if err != nil {
		return nil, err
	}

	defer f.Close()

	var ss []Source

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		fi := strings.Fields(sc.Text())

		if len(fi) == 0 || fi[0][0] == '#' {
			continue
		}

		var d string

		if len(fi) > 1 {
			d = fi[1] // TODO ban absolute path?
		}

		var s Source

		switch {
		case strings.HasPrefix(fi[0], "git+"):
			s, err = newGit(strings.TrimPrefix(fi[0], "git+"), d)
		case strings.Contains(fi[0], "://"):
			s, err = newHTTP(p, fi[0], d)
		default:
			s, err = newFile(p, fi[0], d)
		}

		if err != nil {
			return nil, fmt.Errorf("parse source %s: %w", fi[0], err)
		}

		ss = append(ss, s)
	}

	return ss, sc.Err()
}

// TODO ban ?no-extract
func newGit(s, d string) (*Git, error) {
	u, err := url.Parse(s)

	if err != nil {
		return nil, err
	}

	if u.Scheme == "file" {
		return nil, ErrSourceGitSchemeInvalid
	}

	g := &Git{
		URL: s,
		ed:  d,
		rs: []config.RefSpec{
			"HEAD:HEAD",
		},
	}

	i := strings.LastIndexAny(s, "@#")

	if i < 0 {
		return g, nil
	}

	switch s[i:][0] {
	case '@':
		g.rs = []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:HEAD", s[i+1:])),
		}
	case '#':
		if !plumbing.IsHash(s[i+1:]) {
			return nil, ErrSourceGitHashInvalid
		}

		g.rs = []config.RefSpec{
			config.RefSpec(s[i+1:] + ":HEAD"),
		}
	}

	g.URL = s[:i]
	return g, nil
}

func newHTTP(p *Package, s, d string) (*HTTP, error) {
	u, err := url.Parse(s)

	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrSourceHTTPSchemeInvalid
	}

	v := u.Query()
	_, ok := v["no-extract"]
	v.Del("no-extract")
	u.RawQuery = v.Encode()

	return &HTTP{
		URL: u.String(),
		ed:  d,
		cs:  filepath.Join(p.cfg.SourceDir, p.Name, d, path.Base(u.EscapedPath())),
		ne:  ok,
		pkg: p,
	}, nil
}

func newFile(p *Package, s, d string) (*File, error) {
	for _, s := range []string{
		filepath.Join(p.Path, s),
		filepath.Join(p.cfg.RootDir, s),
	} {
		_, err := os.Stat(s)

		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, err
		}

		return &File{
			Path: s,
			ed:   d,
			pkg:  p,
		}, nil
	}

	return nil, ErrSourceFileNotFound
}

func (g *Git) String() string {
	return path.Base(g.URL)
}

func (h *HTTP) String() string {
	return path.Base(h.URL)
}

func (f *File) String() string {
	return filepath.Base(f.Path)
}
