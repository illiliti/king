package fetch

import (
	"os"

	"github.com/go-git/go-git/v5"
)

// TODO signal handling
// TODO https://github.com/go-git/go-git/pull/58
func GitClone(s, d string) error {
	_, err := git.PlainClone(d, false, &git.CloneOptions{
		URL:          s,
		Depth:        1,
		Progress:     os.Stderr,
		SingleBranch: true,
	})

	return err
}

func GitPull(d string) error {
	r, err := git.PlainOpen(d)

	if err != nil {
		return err
	}

	w, err := r.Worktree()

	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{
		Progress:          os.Stderr,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err == git.ErrRemoteNotFound {
		return nil
	}

	return err
}
