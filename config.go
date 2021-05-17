package king

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/illiliti/king/manifest"
	"golang.org/x/sync/errgroup"
)

// TODO better docs

// Config represents in-memory copy of configuration.
//
// See https://k1sslinux.org/package-manager#4.0
// TODO embed ConfigOptions
type Config struct {
	// Repositories contains list of user-defined repositories.
	Repositories []string

	// DatabaseDir contains path with prepended RootDir to installed packages.
	DatabaseDir string

	// AlternativeDir contains path with prepended RootDir to occurred conflicts.
	AlternativeDir string

	// BinaryDir contains path to pre-built packages.
	BinaryDir string

	// SourceDir contains path to package sources.
	SourceDir string

	// RootDir contains path to real root.
	RootDir string

	db string
	ad string

	// TODO semaphores ?
	// TODO save this info on disk ? /var/db/king/{owners,revdeps}
	ppc uint32
	ppm sync.Mutex
	pp  map[string]*Package

	ddc uint32
	ddm sync.Mutex
	dd  map[string][]string
}

// ConfigOptions allows to configure returned configuration.
type ConfigOptions struct {
	// Repositories contains list of user-defined repositories.
	Repositories []string

	// AlternativeDir contains path with prepended RootDir to occurred conflicts.
	AlternativeDir string

	// DatabaseDir contains path with prepended RootDir to installed packages.
	DatabaseDir string

	// BinaryDir points where pre-built packages will be located.
	BinaryDir string

	// SourceDir points where sources of packages will be downloaded.
	SourceDir string

	// RootDir intended to specify real root of filesystem. Default is '/'.
	// Useful to bootstrap new system.
	RootDir string
}

// NewConfig allocates new instance of configuration.
func NewConfig(co *ConfigOptions) (*Config, error) {
	if err := co.Validate(); err != nil {
		return nil, fmt.Errorf("validate ConfigOptions: %w", err)
	}

	return &Config{
		Repositories:   co.Repositories,
		AlternativeDir: filepath.Join(co.RootDir, co.AlternativeDir),
		DatabaseDir:    filepath.Join(co.RootDir, co.DatabaseDir),
		BinaryDir:      co.BinaryDir,
		SourceDir:      co.SourceDir,
		RootDir:        co.RootDir,
		ad:             co.AlternativeDir,
		db:             co.DatabaseDir,
	}, nil
}

func (c *Config) initOwnedPaths() error {
	if atomic.LoadUint32(&c.ppc) == 1 {
		return nil
	}

	c.ppm.Lock()
	defer c.ppm.Unlock()

	if atomic.LoadUint32(&c.ppc) == 1 {
		return nil
	}

	dd, err := os.ReadDir(c.DatabaseDir)

	if err != nil {
		return err
	}

	var (
		mx sync.Mutex
		eg errgroup.Group
	)

	// internal use only
	var errMultipleOwners = errors.New("multiple owners")

	c.pp = make(map[string]*Package, len(c.pp))

	for _, de := range dd {
		// TODO panic
		if !de.IsDir() {
			continue
		}

		n := de.Name()

		eg.Go(func() error {
			sp, err := NewPackage(c, &PackageOptions{
				Name: n,
				From: Database,
			})

			if err != nil {
				return err
			}

			mf, err := manifest.Open(filepath.Join(sp.Path, "manifest"), os.O_RDONLY)

			if err != nil {
				return err
			}

			defer mf.Close()

			for _, p := range mf.Sort(manifest.NoSort) {
				if strings.HasSuffix(p, "/") { // TODO mark path as directory
					continue
				}

				mx.Lock()
				op, ok := c.pp[p]
				mx.Unlock()

				if ok {
					return fmt.Errorf("parse %s path: %w: [%s %s]", p, errMultipleOwners, op.Name, sp.Name)
				}

				mx.Lock()
				c.pp[p] = sp
				mx.Unlock()
			}

			return nil
		})
	}

	err = eg.Wait()

	// it's better to panic here because DatabaseDir is malformed
	// and we can't reliably operate with the whole Config anymore
	if errors.Is(err, errMultipleOwners) {
		panic(err)
	}

	if err != nil {
		return err
	}

	atomic.StoreUint32(&c.ppc, 1)
	return nil
}

// TODO allow Config.Repositories
func (c *Config) initReverseDependencies() error {
	if atomic.LoadUint32(&c.ddc) == 1 {
		return nil
	}

	c.ddm.Lock()
	defer c.ddm.Unlock()

	if atomic.LoadUint32(&c.ddc) == 1 {
		return nil
	}

	dd, err := os.ReadDir(c.DatabaseDir)

	if err != nil {
		return err
	}

	var (
		mx sync.Mutex
		eg errgroup.Group
	)

	c.dd = make(map[string][]string, len(dd))

	for _, de := range dd {
		// TODO panic
		if !de.IsDir() {
			continue
		}

		n := de.Name()

		eg.Go(func() error {
			sp, err := NewPackage(c, &PackageOptions{
				Name: n,
				From: Database,
			})

			if err != nil {
				return err
			}

			dd, err := sp.Dependencies()

			if errors.Is(err, os.ErrNotExist) {
				return nil
			}

			if err != nil {
				return err
			}

			for _, d := range dd {
				if d.IsMake {
					continue
				}

				mx.Lock()
				c.dd[d.Name] = append(c.dd[d.Name], sp.Name)
				mx.Unlock()
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	atomic.StoreUint32(&c.ddc, 1)
	return nil
}

// ResetOwnedPaths resets cached manifests of installed packages.
func (c *Config) ResetOwnedPaths() {
	c.ppm.Lock()
	defer c.ppm.Unlock()
	atomic.StoreUint32(&c.ppc, 0)
}

// ResetReverseDependencies resets cached reverse dependencies.
func (c *Config) ResetReverseDependencies() {
	c.ddm.Lock()
	defer c.ddm.Unlock()
	atomic.StoreUint32(&c.ddc, 0)
}
