package king

import (
	"os"
	"path/filepath"
	"strconv"
)

// TODO rename to Config ?
type Context struct {
	SysDB  string
	UserDB []string

	CacheDir   string
	SourceDir  string
	LogDir     string
	BinDir     string
	PkgDir     string
	BuildDir   string
	ExtractDir string

	RootDir   string
	ProcessID string

	CompressFormat string

	UserHook string

	HasKeepLog bool
	HasChoice  bool
	HasPrompt  bool
	HasStrip   bool
	HasForce   bool
	HasDebug   bool
}

func NewContext() (*Context, error) {
	cd, err := os.UserCacheDir()
	cp := os.Getenv("KISS_COMPRESS")
	td := os.Getenv("KISS_TMPDIR")
	rd := os.Getenv("KISS_ROOT")
	pd := os.Getenv("KISS_PID")

	if err != nil {
		return nil, err
	}

	cd = filepath.Join(cd, "kiss")

	if cp == "" {
		cp = "gz"
	}

	if pd == "" {
		pd = strconv.Itoa(os.Getpid())
	}

	if rd == "" {
		rd = "/"
	}

	if td == "" {
		td = filepath.Join(cd, "proc", pd)
	}

	return &Context{
		SysDB:          filepath.Join(rd, "var/db/kiss/installed"),
		UserDB:         filepath.SplitList(os.Getenv("KISS_PATH")),
		CacheDir:       cd,
		SourceDir:      filepath.Join(cd, "sources"),
		LogDir:         filepath.Join(cd, "logs"),
		BinDir:         filepath.Join(cd, "bin"),
		PkgDir:         filepath.Join(td, "pkg"),
		BuildDir:       filepath.Join(td, "build"),
		ExtractDir:     filepath.Join(td, "extract"),
		RootDir:        rd,
		ProcessID:      pd,
		CompressFormat: cp,
		UserHook:       os.Getenv("KISS_HOOK"),
		HasKeepLog:     os.Getenv("KISS_KEEPLOG") == "1",
		HasChoice:      os.Getenv("KISS_CHOICE") != "0",
		HasPrompt:      os.Getenv("KISS_PROMPT") != "0",
		HasStrip:       os.Getenv("KISS_STRIP") != "0",
		HasForce:       os.Getenv("KISS_FORCE") == "1",
		HasDebug:       os.Getenv("KISS_DEBUG") == "1",
	}, nil
}
