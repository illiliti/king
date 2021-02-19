package king

import (
	"context"
	"os"
	"os/signal"

	"github.com/go-git/go-git/v5"
)

// Candidate represents candidate for upgrading.
type Candidate struct {
	Name string
}

// Update updates repositories, interates over SysDB and returns non-empty
// slice of Candidate pointers if at least one installed package version differs
// to version available in repositories.
//
// TODO intergrate "kiss-outdated" functionality
func Update(c *Config) ([]*Candidate, error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for _, db := range c.UserDB {
		r, err := git.PlainOpenWithOptions(db, &git.PlainOpenOptions{
			DetectDotGit: true,
		})

		if err != nil {
			return nil, err
		}

		w, err := r.Worktree()

		if err != nil {
			return nil, err
		}

		if err := w.PullContext(ctx, &git.PullOptions{
			Progress:          os.Stderr,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}); err != nil && err != git.NoErrAlreadyUpToDate && err != git.ErrRemoteNotFound {
			return nil, err
		}
	}

	dd, err := os.ReadDir(c.SysDB)

	if err != nil {
		return nil, err
	}

	cc := make([]*Candidate, 0, len(dd))

	// TODO parallelism
	for _, de := range dd {
		sp, err := NewPackageByName(c, Sys, de.Name())

		if err != nil {
			return nil, err
		}

		sv, err := sp.Version()

		if err != nil {
			return nil, err
		}

		up, err := NewPackageByName(c, Usr, de.Name())

		if err != nil {
			continue
		}

		uv, err := up.Version()

		if err != nil {
			return nil, err
		}

		if *sv == *uv {
			continue
		}

		cc = append(cc, &Candidate{
			Name: up.Name,
		})
	}

	return cc, nil
}

func (c *Candidate) String() string {
	return c.Name
}
