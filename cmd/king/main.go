package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/illiliti/king/internal/log"

	"github.com/cornfeedhobo/pflag"
	"github.com/illiliti/king"
)

var version = "BETA"

// TODO KISS_CHOICE KISS_HOOK KISS_COLOR KISS_KEEPLOG KISS_PID

func root(args []string) error {
	cd, err := os.UserCacheDir()

	if err != nil {
		return err
	}

	cd = filepath.Join(cd, "kiss")
	td := os.Getenv("KISS_TMPDIR")

	if td == "" {
		td = filepath.Join(cd, "proc")
	}

	// td = filepath.Join(td, hash.Random(10))
	// TODO KISS_PID
	// TODO use XDG_STATE_HOME for builds
	td = filepath.Join(td, strconv.Itoa(os.Getpid()))

	// XXX be compatible with kiss... only for now
	co := &king.ConfigOptions{
		AlternativeDir: "/var/db/kiss/choices",
		DatabaseDir:    "/var/db/kiss/installed",
	}

	// true, if -v/--version passed to command line
	var fv bool

	// use pflag.ExitOnError to call os.Exit(2) if options are incorrectly specified
	pf := pflag.NewFlagSet("", pflag.ExitOnError)

	// pointer to where to store data, long option, optional short option, default value if option not specified, dummy usage
	pf.StringSliceVar(&co.Repositories, "repository", filepath.SplitList(os.Getenv("KISS_PATH")), "")
	pf.StringVar(&co.BinaryDir, "binary-dir", filepath.Join(cd, "bin"), "")
	pf.StringVar(&co.SourceDir, "source-dir", filepath.Join(cd, "sources"), "")
	pf.StringVar(&co.RootDir, "root-dir", os.Getenv("KISS_ROOT"), "")
	pf.BoolVarP(&fv, "version", "v", false, "")

	// stop parsing options at the first non-option argument.
	// if true, 'king <action> --help' breaks.
	pf.SetInterspersed(false)

	// overwrite usage with our own because i don't like the default one.
	pf.Usage = func() {
		fmt.Fprintln(os.Stderr, kingUsage)
	}

	pf.Parse(args[1:])

	if fv {
		fmt.Println(version)
		return nil
	}

	c, err := king.NewConfig(co)

	if err != nil {
		return err
	}

	switch pf.Arg(0) {
	case "build", "b":
		err = build(c, td, pf.Args())
	case "checksum", "c":
		err = checksum(c, pf.Args())
	case "download", "d":
		err = download(c, pf.Args())
	case "install", "i":
		err = install(c, td, pf.Args())
	case "query", "q":
		err = query(c, pf.Args())
	case "remove", "r":
		err = remove(c, pf.Args())
	case "swap", "s":
		err = swap(c, pf.Args())
	case "update", "u":
		err = update(c, td, pf.Args())
	default:
		pf.Usage()
		os.Exit(2)
	}

	return err
}

func main() {
	if err := root(os.Args); err != nil {
		log.Fatal(err)
	}
}
