package king

import (
	"context"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/henvic/ctxsignal"
	"github.com/illiliti/king/internal/file"
)

// TODO intergrate 'kiss-outdated' functionality

func (c *Config) Update() ([]*Package, error) {
	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

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

	dd, err := file.ReadDirNames(c.SysDB)

	if err != nil {
		return nil, err
	}

	pp := make([]*Package, 0, len(dd))

	// TODO concurrency
	for _, n := range dd {
		sp, err := c.NewPackageByName(Sys, n)

		if err != nil {
			return nil, err
		}

		sv, err := sp.Version()

		if err != nil {
			return nil, err
		}

		up, err := c.NewPackageByName(Usr, n)

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

		pp = append(pp, up)
	}

	return pp, nil
}
