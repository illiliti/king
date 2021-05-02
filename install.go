package king

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/archive"
	"github.com/illiliti/king/internal/cleanup"
	"github.com/illiliti/king/internal/cp"
	"github.com/illiliti/king/internal/hash"

	"github.com/illiliti/king/etcsums"
	"github.com/illiliti/king/manifest"
)

// TODO unit tests
// TODO better docs

var (
	ErrInstallUnmetDependencies = errors.New("target requires other packages")
)

// InstallOptions provides facilities for installing package.
type InstallOptions struct {
	// NoCheckDependencies forcefully installs package without
	// checking if dependencies are met.
	NoCheckDependencies bool

	// OverwriteEtcFiles overwrites /etc/* files without special handling.
	//
	// See https://k1sslinux.org/package-manager#3.3
	OverwriteEtcFiles bool

	// TODO mention that t.Name is appended
	// ExtractDir specifies where pre-built package will be extracted.
	ExtractDir string

	// Debug preserves ExtractDir. Useful for debugging purposes.
	Debug bool
}

// Install deploys pre-built package into system and carefully handles occurred conflicts.
func (t *Tarball) Install(lo *InstallOptions) (*Package, error) {
	if err := lo.Validate(); err != nil {
		return nil, fmt.Errorf("validate InstallOptions: %w", err)
	}

	ed := filepath.Join(lo.ExtractDir, t.Name)
	edp := filepath.Join(ed, t.cfg.db, t.Name)
	pdp := filepath.Join(t.cfg.DatabaseDir, t.Name)

	pmf, pes, err := openManifestEtcsums(pdp, lo.OverwriteEtcFiles, true)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if err := os.MkdirAll(ed, 0777); err != nil {
		return nil, err
	}

	if !lo.Debug {
		defer cleanup.Run(func() error {
			return os.RemoveAll(ed)
		})()
	}

	if err := archive.Extract(t.Path, ed, 0); err != nil {
		return nil, err
	}

	if !lo.NoCheckDependencies {
		if err := unmetDependencies(t, edp); err != nil {
			return nil, err
		}
	}

	emf, err := manifest.Open(filepath.Join(edp, "manifest"), os.O_RDWR)

	if err != nil {
		return nil, err
	}

	if err := handleAlternatives(t, emf, ed); err != nil {
		return nil, err
	}

	defer t.cfg.ResetOwnedPaths()
	defer t.cfg.ResetReverseDependencies()

	for _, i := range emf.Sort(manifest.Directories) {
		if err := install(pes, ed, t.cfg.RootDir, i); err != nil {
			return nil, err
		}
	}

	for _, r := range pmf.Sort(manifest.Files) {
		if emf.Has(r) {
			continue
		}

		if err := remove(pes, t.cfg.RootDir, r); err != nil {
			return nil, err
		}
	}

	return &Package{
		Name: t.Name,
		Path: filepath.Join(t.cfg.db, t.Name),
		cfg:  t.cfg,
	}, nil
}

func openManifestEtcsums(pdp string, ies, cmf bool) (*manifest.Manifest, *etcsums.Etcsums, error) {
	mf, err := manifest.Open(filepath.Join(pdp, "manifest"), os.O_RDONLY)

	if err != nil {
		return nil, nil, err
	}

	if cmf {
		defer mf.Close()
	}

	if ies {
		return mf, nil, nil
	}

	es, err := etcsums.Open(filepath.Join(pdp, "etcsums"), os.O_RDONLY)

	if errors.Is(err, os.ErrNotExist) {
		return mf, nil, nil
	}

	if err != nil {
		return nil, nil, err
	}

	defer es.Close()
	return mf, es, nil
}

// TODO move to dependency.go?
func unmetDependencies(t *Tarball, edp string) error {
	f, err := os.Open(filepath.Join(edp, "depends"))

	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	defer f.Close()

	var pp []string

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		fi := strings.Fields(sc.Text())

		if len(fi) == 0 || fi[0][0] == '#' {
			continue
		}

		p, err := NewPackage(t.cfg, &PackageOptions{
			Name: fi[0],
			From: Database,
		})

		if err != nil {
			pp = append(pp, p.Name)
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	if len(pp) == 0 {
		return nil
	}

	return fmt.Errorf("install package %s: %w: %s", t.Name, ErrInstallUnmetDependencies, pp)
}

func handleAlternatives(t *Tarball, emf *manifest.Manifest, ed string) error {
	pp := emf.Sort(manifest.NoSort)
	cc := make([]string, 0, len(pp))

	for _, p := range pp {
		sp, err := NewPackage(t.cfg, &PackageOptions{
			Path: p,
		})

		if errors.Is(err, ErrPackagePathNotFound) {
			return nil
		}

		if err != nil {
			return err
		}

		if sp.Name == t.Name {
			continue
		}

		cc = append(cc, p)
	}

	if len(cc) == 0 {
		return nil
	}

	if err := os.MkdirAll(filepath.Join(ed, t.cfg.db), 0777); err != nil {
		return err
	}

	for _, c := range cc {
		a := filepath.Join(t.cfg.db, t.Name+strings.ReplaceAll(c, "/", ">"))

		if err := os.Rename(filepath.Join(ed, c), filepath.Join(ed, a)); err != nil {
			return err
		}
	}

	if err := emf.Generate(ed); err != nil {
		return err
	}

	if err := emf.Flush(); err != nil {
		return err
	}

	return emf.Close()
}

func install(es *etcsums.Etcsums, ed, r, p string) error {
	dp := filepath.Join(ed, p)
	rp := filepath.Join(r, p)
	st, err := os.Lstat(dp)

	if err != nil {
		return err
	}

	if st.Mode().IsRegular() && strings.HasPrefix(p, "/etc/") {
		dx, err := hash.Sha256(dp)

		if err != nil {
			return err
		}

		rx, err := hash.Sha256(rp)

		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		if err != nil {
			return err
		}

		// TODO document and recheck https://k1sslinux.org/package-manager#3.3
		switch {
		case rx == dx:
			//
		case !es.Has(rx) && es.Has(dx):
			return nil
		case es.Has(rx) && !es.Has(dx):
			//
		default:
			rp += ".new"
		}
	}

	switch {
	case st.IsDir():
		err = os.MkdirAll(rp, 0777) // TODO umask is locking us to extra chmod()
	case st.Mode().IsRegular():
		return cp.CopyFile(dp, rp)
	case st.Mode()&os.ModeSymlink != 0:
		return cp.CopyLink(dp, rp)
	default:
		// TODO special files, nodes, pipes, etc...
		return nil
	}

	if err != nil {
		return err
	}

	return os.Chmod(rp, st.Mode())
}
