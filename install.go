package king

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/illiliti/king/internal/etcsum"
	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/manifest"
)

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

	err := unmetDependencies(t, filepath.Join(edp, "depends"))

	if !force && err != nil {
		return nil, err
	}

	nm, err := manifest.Open(filepath.Join(edp, "manifest"))

	if err != nil {
		return nil, err
	}

	om, err := manifest.Open(filepath.Join(pdp, "manifest"))

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	defer om.Close()

	oe, err := etcsum.Open(filepath.Join(pdp, "etcsums"))

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	defer oe.Close()
	defer pathsOnce.Reset()
	defer dependenciesOnce.Reset()

	// TODO resolve directory symlinks ?
	for _, i := range nm.Install() {
		sp, err := t.cfg.NewPackageByPath(i)

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

		switch {
		case st.IsDir():
			if err := os.MkdirAll(rp, 0); err != nil {
				return nil, err
			}

			err = os.Chmod(rp, st.Mode()) // TODO
		case st.Mode()&os.ModeSymlink != 0:
			err = file.CopySymlink(dp, rp)
		case st.Mode().IsRegular():
			if strings.HasPrefix(i, "/etc/") {
				dx, err := file.Sha256(dp)

				if err != nil {
					return nil, err
				}

				rx, err := file.Sha256(rp)

				// TODO document and recheck https://k1ss.org/package-manager#3.3
				switch {
				case os.IsNotExist(err):
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

			err = file.CopyFile(dp, rp)
		default:
			continue
		}

		if err != nil {
			return nil, err
		}
	}

	for _, r := range om.Remove() {
		if nm.HasEntry(r) {
			continue
		}

		rp := filepath.Join(t.cfg.RootDir, r)
		st, err := os.Lstat(rp)

		if os.IsNotExist(err) {
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
			dd, err := file.ReadDirNames(rp)

			if err != nil {
				return nil, err
			}

			if len(dd) > 0 {
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

	if os.IsNotExist(err) {
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

		_, err := t.cfg.NewPackageByName(Sys, fi[0])

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

	return fmt.Errorf("install %s: depends on %s", t.Name, strings.Join(pp, ", "))
}
