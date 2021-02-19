package king

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/etcsum"
	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/manifest"
)

// Install installs given package to the system.
//
// If force equal false, Install checks if packages declared in
// depends file (if exists) are installed and returns error if not.
// Otherwise, this check is entirely skipped.
func (t *Tarball) Install(force bool) (*Package, error) {
	ed := filepath.Join(t.cfg.ExtractDir, t.Name)
	edp := filepath.Join(ed, InstalledDir, t.Name)
	pdp := filepath.Join(t.cfg.SysDB, t.Name)

	if err := os.MkdirAll(ed, 0777); err != nil {
		return nil, err
	}

	// TODO cleanup on signal
	if !t.cfg.HasDebug {
		defer os.RemoveAll(ed)
	}

	if err := file.Unarchive(t.Path, ed, 0); err != nil {
		return nil, err
	}

	if !force {
		err := unmetDependencies(t, filepath.Join(edp, "depends"))

		if err != nil {
			return nil, err
		}
	}

	nm, err := manifest.Open(filepath.Join(edp, "manifest"))

	if err != nil {
		return nil, err
	}

	om, err := manifest.Open(filepath.Join(pdp, "manifest"))

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	defer om.Close()

	oe, err := etcsum.Open(filepath.Join(pdp, "etcsums"))

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	defer oe.Close()

	// TODO resolve directory symlinks ?
	for _, i := range nm.Install() {
		sp, err := NewPackageByPath(t.cfg, i)

		if err != nil {
			continue
		}

		if sp.Name == t.Name {
			continue
		}

		a := filepath.Join(ChoicesDir, t.Name+strings.ReplaceAll(i, "/", ">"))

		if err := os.MkdirAll(filepath.Join(ed, ChoicesDir), 0777); err != nil {
			return nil, err
		}

		if err := os.Rename(filepath.Join(ed, i), filepath.Join(ed, a)); err != nil {
			return nil, err
		}
	}

	if err := nm.Generate(ed); err != nil {
		return nil, err
	}

	if err := nm.Flush(); err != nil {
		return nil, err
	}

	if err := nm.Close(); err != nil {
		return nil, err
	}

	defer pathsOnce.Reset()
	defer dependenciesOnce.Reset()

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	// TODO incremental installation

	for _, i := range nm.Install() {
		rp := filepath.Join(t.cfg.RootDir, i)
		dp := filepath.Join(ed, i)
		st, err := os.Lstat(dp)

		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(i, "/etc/") && !strings.HasSuffix(i, "/") {
			dx, err := file.Sha256(dp)

			if err != nil {
				return nil, err
			}

			rx, err := file.Sha256(rp)

			// TODO document and recheck https://kiss.armaanb.net/package-manager#3.3
			switch {
			case errors.Is(err, fs.ErrNotExist):
				//
			case err != nil:
				return nil, err
			case rx == dx:
				//
			case !oe.HasEntry(rx) && oe.HasEntry(dx):
				continue
			case oe.HasEntry(rx) && !oe.HasEntry(dx):
				//
			default:
				rp += ".new"
			}
		}

		switch {
		case st.IsDir():
			err = os.MkdirAll(rp, 0)
		case st.Mode().IsRegular():
			err = file.CopyFile(dp, rp)
		case st.Mode()&fs.ModeSymlink != 0:
			err = file.CopySymlink(dp, rp)
		default:
			continue
		}

		if err != nil {
			return nil, err
		}

		if err := os.Chmod(rp, st.Mode()); err != nil {
			return nil, err
		}
	}

	for _, r := range om.Remove() {
		if nm.HasEntry(r) {
			continue
		}

		rp := filepath.Join(t.cfg.RootDir, r)
		st, err := os.Lstat(rp)

		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return nil, err
		}

		switch {
		case st.Mode().IsRegular() && strings.HasPrefix(r, "/etc/"):
			x, err := file.Sha256(rp)

			if err != nil {
				return nil, err
			}

			if !oe.HasEntry(x) {
				continue
			}
		case st.IsDir():
			f, err := os.Open(rp)

			if err != nil {
				return nil, err
			}

			_, err = f.ReadDir(1)

			if err := f.Close(); err != nil {
				return nil, err
			}

			if !errors.Is(err, io.EOF) {
				continue
			}
		}

		if err := os.Remove(rp); err != nil {
			return nil, err
		}
	}

	// TODO verify installation ?

	return &Package{
		Name: t.Name,
		Path: pdp,
		cfg:  t.cfg,
	}, nil
}

func unmetDependencies(t *Tarball, d string) error {
	f, err := os.Open(d)

	if errors.Is(err, fs.ErrNotExist) {
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

		_, err := NewPackageByName(t.cfg, Sys, fi[0])

		if err != nil {
			pp = append(pp, fi[0])
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}

	if len(pp) == 0 {
		return nil
	}

	return fmt.Errorf("install %s: depends on %s", t.Name, pp)
}
