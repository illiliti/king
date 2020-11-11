package king

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func (c *Context) RunRepoHook(n, p string) error {
	h := filepath.Join(c.SysDB, p, n)
	st, err := os.Stat(h)

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	if m := st.Mode(); !m.IsDir() && m&0111 != 0 {
		cmd := exec.Command(h)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout

		return cmd.Run()
	}

	b, err := ioutil.ReadFile(h)

	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func (c *Context) RunUserHook(t, p, d string) error {
	if c.UserHook == "" {
		return nil
	}

	cmd := exec.Command(c.UserHook)
	cmd.Env = append(os.Environ(), "TYPE="+t, "PKG="+p, "DEST="+d)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd.Run()
}
