package cleanup

import (
	"os"
	"os/signal"
)

// TODO unit tests

func Run(f func() error) func() {
	cs := make(chan os.Signal, 1)

	go func() {
		signal.Notify(cs, os.Interrupt, os.Kill)
		defer signal.Stop(cs)

		s, ok := <-cs

		if !ok {
			return
		}

		if err := f(); err != nil {
			panic(err)
		}

		// https://pubs.opengroup.org/onlinepubs/9699919799/xrat/V4_xcu_chap02.html#tag_18_08_02
		c := 128

		// TODO according to POSIX/ISO C, signal code is implementation-defined
		// and we need more portable way to exit with corresponding signal code
		switch s {
		case os.Kill:
			c += 15
		case os.Interrupt:
			c += 2
		}

		os.Exit(c)
	}()

	return func() {
		defer close(cs)

		if err := f(); err != nil {
			panic(err)
		}
	}
}
