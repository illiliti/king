package king

import (
	"bufio"
	"context"
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/illiliti/king/internal/archive"
	"github.com/illiliti/king/internal/cp"

	"github.com/illiliti/king/etcsums"
	"github.com/illiliti/king/manifest"
	"golang.org/x/sync/errgroup"
)

// TODO unit tests
// TODO better docs

var systemLibrary = map[string]bool{
	"c":       true,
	"m":       true,
	"dl":      true,
	"rt":      true,
	"xnet":    true,
	"util":    true,
	"trace":   true,
	"crypt":   true,
	"resolv":  true,
	"pthread": true,
}

// BuildOptions provides facilities for building package.
type BuildOptions struct {
	// NoStripBinaries disables discarding unnecessary symbols
	// from binaries and libraries.
	//
	// NoStripBinaries bool

	// Output points where build log will be written.
	Output io.Writer

	// Compression defines which compression format will be used to create tarball.
	//
	// TODO allow none format
	// Valid formats are sz, br, gz, xz, zst, bz2, lz4.
	Compression string

	// TODO mention that p.Name is appended
	// BuildDir specifies where build will happen.
	BuildDir string

	// TODO mention that p.Name is appended
	// PackageDir specifies where build script places package files
	// to turn them into tarball later.
	PackageDir string

	// Debug preserves BuildDir and PackageDir. Useful for debugging purposes.
	Debug bool
}

// Build builds package and turns it into installable tarball.
//
// See https://k1sslinux.org/package-system#2.0
func (p *Package) Build(bo *BuildOptions) (*Tarball, error) {
	return p.BuildContext(context.Background(), bo)
}

func (p *Package) BuildContext(ctx context.Context, bo *BuildOptions) (*Tarball, error) {
	if err := bo.Validate(); err != nil {
		return nil, fmt.Errorf("validate BuildOptions: %w", err)
	}

	v, err := p.Version()

	if err != nil {
		return nil, err
	}

	ss, err := p.Sources()

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	bd := filepath.Join(bo.BuildDir, p.Name)
	pd := filepath.Join(bo.PackageDir, p.Name)
	pdp := filepath.Join(pd, p.cfg.db, p.Name)

	for _, d := range []string{bd, pd} {
		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}
	}

	if !bo.Debug {
		defer os.RemoveAll(bd)
		defer os.RemoveAll(pd)
	}

	// TODO add a way to reuse bo.BuildDir, i.e skip this loop if bo.ReuseBuildDir defined
	for _, s := range ss {
		d := filepath.Join(bd, s.ExtractDir())

		// TODO drop
		if err := os.MkdirAll(d, 0777); err != nil {
			return nil, err
		}

		if err := s.Extract(d); err != nil {
			return nil, err
		}
	}

	cmd := exec.CommandContext(ctx, filepath.Join(p.Path, "build"), pd, v.Version)
	cmd.Stdout = bo.Output
	cmd.Stderr = bo.Output
	cmd.Dir = bd

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("execute build: %w", err)
	}

	if err := removeGarbage(pd); err != nil {
		return nil, fmt.Errorf("remove garbage: %w", err)
	}

	// TODO strip binaries
	// if err := stripBinaries(pd); err != nil {
	// 	return nil, fmt.Errorf("strip binaries: %w", err)
	// }

	if err := cp.CopyDir(p.Path, pdp); err != nil {
		return nil, fmt.Errorf("copy database: %w", err)
	}

	if err := updateDepends(p, pd, pdp); err != nil {
		return nil, fmt.Errorf("update dependencies: %w", err)
	}

	if err := createEtcsums(pd, pdp); err != nil {
		return nil, fmt.Errorf("create etcsums: %w", err)
	}

	if err := updateManifest(pd, pdp); err != nil {
		return nil, fmt.Errorf("update manifest: %w", err)
	}

	t := &Tarball{
		Name: p.Name,
		Path: filepath.Join(p.cfg.BinaryDir, p.Name+"@"+
			v.Version+"-"+v.Release+".tar."+bo.Compression),
		cfg: p.cfg,
	}

	if err := archive.Create(pd, t.Path); err != nil {
		return nil, err
	}

	return t, nil
}

func removeGarbage(pd string) error {
	err := filepath.WalkDir(filepath.Join(pd, "usr/lib"), func(p string, de os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !de.Type().IsRegular() {
			return nil
		}

		if de.Name() != "charset.alias" && filepath.Ext(de.Name()) != ".la" {
			return nil
		}

		return os.Remove(p)
	})

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func updateDepends(bp *Package, pd, pdp string) error {
	var (
		mx sync.Mutex
		eg errgroup.Group
	)

	dd := make(map[string]bool)

	// TODO only /usr/{lib,bin}
	err := filepath.WalkDir(pd, func(p string, de os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !de.Type().IsRegular() {
			return nil
		}

		f, err := elf.Open(p)

		if err != nil {
			return nil
		}

		defer f.Close()

		ll, err := f.ImportedLibraries()

		if err != nil {
			return nil
		}

		for _, l := range ll {
			l := l // HACK

			eg.Go(func() error {
				i := strings.Index(l, ".")

				if i > 3 && systemLibrary[l[3:i]] {
					return nil
				}

				// TODO stop hardcoding /usr/lib
				// use LD_LIBRARY_PATH, DT_RUNPATH, DT_RPATH, /etc/ld.so.conf, /etc/ld-musl-<arch>.path
				// https://cgit.uclibc-ng.org/cgi/cgit/uclibc-ng.git/tree/utils/ldd.c?id=672a303852353ba9299f6f50190fca8b3abe4c1d#n489
				sp, err := NewPackage(bp.cfg, &PackageOptions{
					Path: filepath.Join("/usr/lib", l),
				})

				if errors.Is(err, ErrPackagePathNotFound) {
					return nil
				}

				if err != nil {
					return err
				}

				if sp.Name == bp.Name {
					return nil
				}

				mx.Lock()
				defer mx.Unlock()

				if _, ok := dd[sp.Name]; !ok {
					dd[sp.Name] = false
				}

				return nil
			})
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if len(dd) == 0 {
		return nil
	}

	f, err := os.OpenFile(filepath.Join(pdp, "depends"), os.O_CREATE|os.O_RDWR, 0666)

	if err != nil {
		return err
	}

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		fi := strings.Fields(sc.Text())

		if len(fi) == 0 || fi[0][0] == '#' {
			continue
		}

		if _, ok := dd[fi[0]]; !ok {
			dd[fi[0]] = len(fi) > 1 && fi[1] == "make"
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	if err := f.Truncate(0); err != nil {
		return err
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	w := bufio.NewWriter(f)

	for n, m := range dd {
		w.WriteString(n)

		if m {
			w.WriteString(" make")
		}

		w.WriteByte('\n')
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return f.Close()
}

func createEtcsums(pd, pdp string) error {
	// XXX
	st, err := os.Lstat(filepath.Join(pd, "etc"))

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	if !st.IsDir() {
		return nil
	}

	es, err := etcsums.Create(filepath.Join(pdp, "etcsums"))

	if err != nil {
		return err
	}

	// XXX
	if err := es.Generate(filepath.Join(pd, "etc")); err != nil {
		return err
	}

	if err := es.Flush(); err != nil {
		return err
	}

	return es.Close()
}

func updateManifest(pd, pdp string) error {
	mf, err := manifest.Create(filepath.Join(pdp, "manifest"))

	if err != nil {
		return err
	}

	if err := mf.Generate(pd); err != nil {
		return err
	}

	if err := mf.Flush(); err != nil {
		return err
	}

	return mf.Close()
}
