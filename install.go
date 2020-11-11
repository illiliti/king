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
)

// TODO return (*Package, error) ?
// TODO signal handling and cleanup
func (t *Tarball) Install(force bool) error {
	ed := filepath.Join(t.context.ExtractDir, t.Name)

	if err := os.MkdirAll(ed, 0777); err != nil {
		return err
	}

	if err := file.ExtractArchive(t.Path, ed, 0); err != nil {
		return err
	}

	pdp := filepath.Join(t.context.SysDB, t.Name)
	edp := filepath.Join(ed, "var/db/kiss/installed", t.Name)

	npp, err := readManifest(filepath.Join(edp, "manifest"))

	if err != nil {
		return err
	}

	opp, err := readManifest(filepath.Join(pdp, "manifest"))

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// ee, err := readEtcsums(filepath.Join(pdp, "etcsums"))

	// if err != nil && !os.IsNotExist(err) {
	// 	return err
	// }

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

			_, err := t.context.NewPackage(fi[0], SysDB)

			if err != nil {
				return err
			}
		}

		return sc.Err()
	}

	if !force {
		for _, p := range npp {
			_, err := os.Lstat(filepath.Join(ed, p))

			if err != nil {
				return err
			}
		}

		if err := checkDepends(); err != nil {
			return err
		}
	}

	mpp := make(map[string]bool, len(npp))

	for _, p := range npp {
		d, n := filepath.Split(p)

		if n == "" {
			continue
		}

		s, err := filepath.EvalSymlinks(filepath.Join(t.context.RootDir, d))

		if err != nil {
			if os.IsNotExist(err) {
				mpp[p] = true
				continue // TODO simplify
			}

			return err
		}

		// FIXME RootDir can be symlink. in that case TrimPrefix will not work
		// KISS don't care about it https://github.com/kisslinux/kiss/blob/1ae8340e49d59edbed54f769b971e09a8e934c45/kiss#L816
		md := "/" + strings.TrimPrefix(s, t.context.RootDir)
		mpp[filepath.Join(md, n)] = true
	}

	var cpp []string

	findConflicts := func(n string) error {
		f, err := os.Open(filepath.Join(t.context.SysDB, n, "manifest"))

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

	dd, err := file.ReadDirNames(t.context.SysDB)

	if err != nil {
		return err
	}

	for _, n := range dd {
		if n == t.Name {
			continue
		}

		if err := findConflicts(n); err != nil {
			return err
		}
	}

	if len(cpp) > 0 {
		if !t.context.HasChoice {
			return fmt.Errorf("package %s conflicts with other package", t.Name)
		}

		cd := filepath.Join(ed, "var/db/kiss/choices")

		if err := os.MkdirAll(cd, 0777); err != nil {
			return err
		}

		for _, p := range cpp {
			cp := filepath.Join(cd, t.Name+strings.ReplaceAll(p, "/", ">"))

			if err := os.Rename(filepath.Join(ed, p), cp); err != nil {
				return err
			}
		}

		npp, err = generateManifest(ed)

		if err != nil {
			return err
		}

		if err := saveManifest(npp, filepath.Join(edp, "manifest")); err != nil {
			return err
		}
	}

	if err := t.context.RunUserHook("pre-install", t.Name, ed); err != nil {
		return err
	}

	signal.Ignore(os.Interrupt)
	defer signal.Reset(os.Interrupt)

	sort.Strings(npp)

	for _, p := range npp {
		dp := filepath.Join(ed, p)
		st, err := os.Lstat(dp)

		if err != nil {
			return err
		}

		m := st.Mode()
		rp := filepath.Join(t.context.RootDir, p)

		if m.IsRegular() && strings.HasPrefix(p, "/etc/") {
			// TODO https://k1ss.org/package-manager#3.3
		}

		if err := installFile(dp, rp, m); err != nil {
			return err
		}
	}

	for _, p := range opp {
		if mpp[p] {
			continue
		}

		rp := filepath.Join(t.context.RootDir, p)
		st, err := os.Stat(rp)

		if err != nil {
			return err
		}

		if err := removeFile(rp, st.Mode()); err != nil {
			return err
		}
	}

	// TODO verify installation ?

	signal.Reset(os.Interrupt)

	if err := t.context.RunRepoHook("post-install", t.Name); err != nil {
		return err
	}

	return t.context.RunUserHook("post-install", t.Name, pdp)
}
