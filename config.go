package king

import (
	"os"
	"path/filepath"
	"strconv"
)

// TODO rename ?
const (
	InstalledDir = "/var/db/kiss/installed"
	ChoicesDir   = "/var/db/kiss/choices"

	// BaseDir        = "/var/db/king"
	// BaseDir        = "/var/lib/king"
	// DatabaseDir    = BaseDir + "/database"
	// AlternativeDir = BaseDir + "/alternative"
	// SourceDir      = BaseDir + "/source"
	// SourceDir      = "/var/tmp/king/source"
	// ...
)

// TODO use config file(toml, yaml, sr.ht/~emersion/go-scfg, github.com/go-ini/ini)
// instead of environment variables ?
type Config struct {
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

// TODO clean
func NewConfig() (*Config, error) {
	cd, err := os.UserCacheDir()
	cp := os.Getenv("KISS_COMPRESS")
	td := os.Getenv("KISS_TMPDIR")
	rd := os.Getenv("KISS_ROOT")
	pd := os.Getenv("KISS_PID")

	if err != nil {
		return nil, err
	}

	if rd != "" {
		rd, err = filepath.EvalSymlinks(rd)

		if err != nil {
			return nil, err
		}
	}

	cd = filepath.Join(cd, "kiss")

	if cp == "" {
		cp = "gz"
	}

	if pd == "" {
		pd = strconv.Itoa(os.Getpid())
	}

	if td == "" {
		td = filepath.Join(cd, "proc", pd)
	}

	return &Config{
		SysDB:          filepath.Join(rd, InstalledDir),
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
