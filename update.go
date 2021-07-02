package king

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"golang.org/x/sync/errgroup"
)

// TODO unit tests
// TODO better docs

// UpdateOptions provides facilities for updating repositories.
type UpdateOptions struct {
	// NoUpdateRepositories disables updating repositories.
	NoUpdateRepositories bool

	// ContinueOnError ignores possible errors during updating repositories.
	ContinueOnError bool

	// ExcludePackages ignores update for specified packages.
	ExcludePackages []string
}

// type Candidate struct {
// 	*Package

// 	ov *Version
// 	nv *Version
// }

// Update optionally updates repositories and parses candidates for upgrade.
func Update(c *Config, uo *UpdateOptions) ([]*Package, error) {
	return UpdateContext(context.Background(), c, uo)
}

func UpdateContext(ctx context.Context, c *Config, uo *UpdateOptions) ([]*Package, error) {
	if !uo.NoUpdateRepositories {
		err := updateRepositories(ctx, c.Repositories)

		if !uo.ContinueOnError && err != nil {
			return nil, fmt.Errorf("update repositories: %w", err)
		}
	}

	epp := make(map[string]bool, len(uo.ExcludePackages))

	for _, n := range uo.ExcludePackages {
		if n != "" {
			epp[n] = true
		}
	}

	dd, err := os.ReadDir(c.DatabaseDir)

	if err != nil {
		return nil, err
	}

	var (
		mx sync.Mutex
		eg errgroup.Group
	)

	pp := make([]*Package, 0, len(dd))

	for _, de := range dd {
		n := de.Name()

		if epp[n] {
			continue
		}

		eg.Go(func() error {
			up, err := NewPackage(c, &PackageOptions{
				Name: n,
				From: Repository,
			})

			if errors.Is(err, ErrPackageNameNotFound) {
				return nil
			}

			if err != nil {
				return err
			}

			uv, err := up.Version()

			if err != nil {
				return err
			}

			sp, err := NewPackage(c, &PackageOptions{
				Name: n,
				From: Database,
			})

			if err != nil {
				return err
			}

			sv, err := sp.Version()

			if err != nil {
				return err
			}

			if *sv == *uv {
				return nil
			}

			mx.Lock()
			defer mx.Unlock()

			pp = append(pp, up)
			return nil
		})
	}

	return pp, eg.Wait()
}

func updateRepositories(ctx context.Context, rr []string) error {
	for _, d := range rr {
		rp, err := filepath.EvalSymlinks(d)

		if err != nil {
			return err
		}

		r, err := git.PlainOpenWithOptions(rp, &git.PlainOpenOptions{
			DetectDotGit: true,
		})

		if err != nil {
			return err
		}

		w, err := r.Worktree()

		if err != nil {
			return err
		}

		err = w.PullContext(ctx, &git.PullOptions{
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if errors.Is(err, git.NoErrAlreadyUpToDate) || errors.Is(err, git.ErrRemoteNotFound) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}
