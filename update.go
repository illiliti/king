package king

import (
	"context"
	"os"
	"os/signal"
	"sync"

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

	var mx sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(dd))
	cc := make([]*Candidate, 0, len(dd))

	for _, de := range dd {
		go func(n string) {
			defer wg.Done()

			up, err := NewPackageByName(c, Usr, n)

			if err != nil {
				return
			}

			uv, err := up.Version()

			if err != nil {
				return
			}

			sp, err := NewPackageByName(c, Sys, n)

			if err != nil {
				return
			}

			sv, err := sp.Version()

			if err != nil {
				return
			}

			if *sv == *uv {
				return
			}

			mx.Lock()
			defer mx.Unlock()

			cc = append(cc, &Candidate{
				Name: up.Name,
			})
		}(de.Name())
	}

	wg.Wait()
	return cc, nil
}

func (c *Candidate) String() string {
	return c.Name
}
