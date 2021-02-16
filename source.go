package king

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// TODO Source as interface ?
// type Source interface {
// 	Prepare(d string) error
//  Download(d string) error
//  Verify() error
//  ...
// }

type Protocol interface {
	Prepare(d string) error
}

type Source struct {
	Protocol
	DestinationDir string
}

type Git struct {
	URL string
}

type HTTP struct {
	URL          string
	Path         string
	HasNoExtract bool

	pkg *Package
}

type File struct {
	Path string

	pkg *Package
}

func (p *Package) Sources() ([]*Source, error) {
	f, err := os.Open(filepath.Join(p.Path, "sources"))

	if err != nil {
		return nil, err
	}

	defer f.Close()

	var ss []*Source

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		fi := strings.Fields(sc.Text())

		if len(fi) == 0 || fi[0][0] == '#' {
			continue
		}

		s := new(Source)

		if len(fi) > 1 {
			s.DestinationDir = fi[1]
		}

		switch {
		case strings.HasPrefix(fi[0], "git+"):
			s.Protocol, err = newGit(strings.TrimPrefix(fi[0], "git+"))
		case strings.Contains(fi[0], "://"):
			s.Protocol, err = newHTTP(p, fi[0], s.DestinationDir)
		default:
			s.Protocol, err = newFile(p, fi[0])
		}

		if err != nil {
			return nil, err
		}

		ss = append(ss, s)
	}

	return ss, sc.Err()
}

func newGit(s string) (*Git, error) {
	if strings.ContainsAny(s, "@#") {
		return nil, fmt.Errorf("source %s: unsupported branch/commit")
	}

	return &Git{
		URL: s,
	}, nil
}

// TODO newFile if source has no-extract
func newHTTP(p *Package, s, d string) (*HTTP, error) {
	u, err := url.Parse(s)

	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("source %s: unsupported protocol", s)
	}

	return &HTTP{
		URL:          s,
		Path:         filepath.Join(p.cfg.SourceDir, p.Name, d, filepath.Base(u.Path)),
		HasNoExtract: u.Query()["no-extract"] != nil, // TODO
		pkg:          p,
	}, nil
}

func newFile(p *Package, s string) (*File, error) {
	for _, s := range []string{
		filepath.Join(p.Path, s),
		filepath.Join(p.cfg.RootDir, s),
	} {
		if _, err := os.Stat(s); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}

		return &File{
			Path: s,
			pkg:  p,
		}, nil
	}

	return nil, fmt.Errorf("source %s: not found", s)
}
