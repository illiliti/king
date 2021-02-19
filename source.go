package king

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
)

// Source abstracts various types of sources that has ExtractDir method, which
// returns (relative to the build directory) directory where sources should be
// placed, and Extract method, which unpacks/downloads sources to the specified
// directory.
//
// See https://kiss.armaanb.net/package-system#4.0
type Source interface {
	ExtractDir() string
	Extract(d string) error
}

// HTTP represents http source.
type HTTP struct {
	u  string
	p  string
	d  string
	ne bool

	pkg *Package
}

// File represents absolute/relative file source.
type File struct {
	p  string
	d  string
	ia bool

	pkg *Package
}

// Git represents git source.
type Git struct {
	u string
	d string
}

// Sources returns slice of Source interfaces for a given package.
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
			d = fi[1]
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
			return nil, err
		}

		ss = append(ss, s)
	}

	return ss, sc.Err()
}

func newGit(s, d string) (*Git, error) {
	if strings.ContainsAny(s, "@#") {
		return nil, fmt.Errorf("source %s: unsupported branch/commit")
	}

	return &Git{
		u: s,
		d: d,
	}, nil
}

func newHTTP(p *Package, s, d string) (*HTTP, error) {
	u, err := url.Parse(s)

	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("source %s: unsupported protocol", s)
	}

	_, ok := u.Query()["no-extract"]

	return &HTTP{
		u:   s,
		p:   filepath.Join(p.cfg.SourceDir, p.Name, d, filepath.Base(u.Path)),
		ne:  ok,
		pkg: p,
	}, nil
}

func newFile(p *Package, s, d string) (*File, error) {
	if !fs.ValidPath(s) {
		return nil, fmt.Errorf("source %s: invalid", s)
	}

	for _, s := range []string{
		filepath.Join(p.Path, s),
		filepath.Join(p.cfg.RootDir, s),
	} {
		_, err := os.Stat(s)

		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, err
		}

		_, err = archiver.ByExtension(s)

		return &File{
			p:   s,
			d:   d,
			ia:  err == nil,
			pkg: p,
		}, nil
	}

	return nil, fmt.Errorf("source %s: not found", s)
}

func (g *Git) String() string {
	return g.u
}

func (h *HTTP) String() string {
	return h.u
}

func (f *File) String() string {
	return f.p
}
