package king

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"

	"github.com/illiliti/king/internal/file"
	"github.com/illiliti/king/internal/skel"
	"github.com/illiliti/king/internal/skel/manifest"
)

// TODO cleanup on signal
func (t *Tarball) Install(force bool) (*Package, error) {
	ed := filepath.Join(t.cfg.ExtractDir, t.Name)

	if err := os.MkdirAll(ed, 0777); err != nil {
		return nil, err
	}

	if !t.cfg.HasDebug {
		defer os.RemoveAll(ed)
	}

	if err := file.ExtractArchive(t.Path, ed, 0); err != nil {
		return nil, err
	}

	pdp := filepath.Join(t.cfg.SysDB, t.Name)
	edp := filepath.Join(ed, InstalledDir, t.Name)

	npp, err := skel.Slice(filepath.Join(edp, "manifest"))

	if err != nil {
		return nil, err
	}

	opp, err := skel.Slice(filepath.Join(pdp, "manifest"))

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	oee, err := skel.Map(filepath.Join(pdp, "etcsums"))

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// TODO refactor ?
	checkDepends := func() error {
		f, err := os.Open(filepath.Join(edp, "depends"))

		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}

			return err
		}

		defer f.Close()

		sc := bufio.NewScanner(f)

		for sc.Scan() {
			fi := strings.Fields(sc.Text())

			if len(fi) == 0 || fi[0][0] == '#' {
				continue
			}

			_, err := t.cfg.NewPackage(fi[0], Sys)

			if err != nil {
				return err
			}
		}

		return sc.Err()
	}

	if !force {
		for _, p := range npp {
			if _, err := os.Lstat(filepath.Join(ed, p)); err != nil {
				return nil, err
			}
		}

		if err := checkDepends(); err != nil {
			return nil, err
		}
	}

	// TODO add dirs too
	mpp := make(map[string]bool, len(npp))

	for _, p := range npp {
		d, n := filepath.Split(p)

		if n == "" {
			continue
		}

		s, err := filepath.EvalSymlinks(filepath.Join(t.cfg.RootDir, d))

		if err != nil {
			if os.IsNotExist(err) {
				s = d
			} else {
				return nil, err
			}
		}

		mpp[filepath.Join(strings.TrimPrefix(s, t.cfg.RootDir), n)] = true
	}

	var cpp []string

	// TODO use Owner()
	findConflicts := func(n string) error {
		f, err := os.Open(filepath.Join(t.cfg.SysDB, n, "manifest"))

		if err != nil {
			return err
		}

		defer f.Close()

		sc := bufio.NewScanner(f)

		for sc.Scan() {
			p := sc.Text()

			if mpp[p] {
				cpp = append(cpp, p)
			}
		}

		return sc.Err()
	}

	dd, err := file.ReadDirNames(t.cfg.SysDB)

	if err != nil {
		return nil, err
	}

	for _, n := range dd {
		if n == t.Name {
			continue
		}

		if err := findConflicts(n); err != nil {
			return nil, err
		}
	}

	// TODO cleanup
	if len(cpp) > 0 {
		if !t.cfg.HasChoice {
			return nil, fmt.Errorf("package %s conflicts with other package", t.Name)
		}

		cd := filepath.Join(ed, ChoicesDir)

		if err := os.MkdirAll(cd, 0777); err != nil {
			return nil, err
		}

		for _, p := range cpp {
			n := t.Name + strings.ReplaceAll(p, "/", ">")

			if mpp[p] {
				delete(mpp, p)
				mpp[filepath.Join(ChoicesDir, n)] = true
			}

			cp := filepath.Join(cd, n)

			if err := os.Rename(filepath.Join(ed, p), cp); err != nil {
				return nil, err
			}
		}

		f, err := os.Create(filepath.Join(edp, "manifest"))

		if err != nil {
			return nil, err
		}

		npp, err = manifest.Generate(ed)

		if err != nil {
			return nil, err
		}

		if err := skel.Save(npp, f); err != nil {
			return nil, err
		}

		if err := f.Close(); err != nil {
			return nil, err
		}
	}

	if err := t.cfg.RunUserHook("pre-install", t.Name, ed); err != nil {
		return nil, err
	}

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	sort.Strings(npp)

	for _, p := range npp {
		dp := filepath.Join(ed, p)
		st, err := os.Lstat(dp)

		if err != nil {
			return nil, err
		}

		rp := filepath.Join(t.cfg.RootDir, p)

		switch m := st.Mode(); {
		case m.IsDir():
			if err := os.MkdirAll(rp, 0777); err != nil {
				return nil, err
			}

			err = os.Chmod(rp, m)
		case m.IsRegular():
			if strings.HasPrefix(p, "/etc/") {
				// TODO https://k1ss.org/package-manager#3.3
			}

			err = file.CopyFile(dp, rp)
		case m&os.ModeSymlink != 0:
			err = file.CopySymlink(dp, rp)
		}

		if err != nil {
			return nil, err
		}
	}

	for _, r := range opp {
		if mpp[r] {
			continue
		}

		rp := filepath.Join(t.cfg.RootDir, r)
		st, err := os.Lstat(rp)

		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}

		switch m := st.Mode(); {
		case m.IsRegular() && strings.HasPrefix(r, "/etc/"):
			h, err := file.Sha256Sum(rp)

			if err != nil {
				return nil, err
			}

			if len(oee) > 0 && !oee[h] {
				continue
			}
		case m.IsDir():
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

	signal.Reset(os.Interrupt)

	if err := t.cfg.RunRepoHook("post-install", t.Name); err != nil {
		return nil, err
	}

	return &Package{
		Name: t.Name,
		Path: pdp,
		cfg:  t.cfg,
	}, t.cfg.RunUserHook("post-install", t.Name, pdp)
}
