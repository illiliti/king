package king

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// TODO better docs

var compressFormat = map[string]bool{
	"sz":  true,
	"br":  true,
	"gz":  true,
	"xz":  true,
	"zst": true,
	"bz2": true,
	"lz4": true,
}

// type ValidateError struct { // TODO

// }

var (
	ErrAlternativePathDirectory   = errors.New("Path cannot be a directory")
	ErrAlternativePathNotAbsolute = errors.New("Path must be absolute")
	ErrAlternativePathEmpty       = errors.New("Path must be set")

	ErrBuildCompressionNotSupported = errors.New("Compression contains an unsupported format")
	ErrBuildPackageDirEmpty         = errors.New("PackageDir must be set")
	ErrBuildBuildDirEmpty           = errors.New("BuildDir must be set")

	ErrConfigRootDirNotDirectory = errors.New("RootDir must be a directory")
	ErrConfigAlternativeDirEmpty = errors.New("AlternativeDir must be set")
	ErrConfigDatabaseDirEmpty    = errors.New("DatabaseDir must be set")
	ErrConfigSourceDirEmpty      = errors.New("SourceDir must be set")
	ErrConfigBinaryDirEmpty      = errors.New("BinaryDir must be set")

	ErrInstallExtractDirEmpty = errors.New("ExtractDir must be set")

	ErrPackagePathNameExclusive = errors.New("Path and Name are mutually exclusive")
)

func (ao *AlternativeOptions) Validate() error {
	if ao.Path == "" {
		return ErrAlternativePathEmpty
	}

	if !filepath.IsAbs(ao.Path) {
		return ErrAlternativePathNotAbsolute
	}

	st, err := os.Lstat(ao.Path)

	if err != nil {
		return err
	}

	if !st.IsDir() {
		return nil
	}

	return ErrAlternativePathDirectory
}

func (bo *BuildOptions) Validate() error {
	if bo.Compression == "" {
		bo.Compression = "gz"
	}

	if !compressFormat[bo.Compression] {
		return ErrBuildCompressionNotSupported
	}

	if bo.BuildDir == "" {
		return ErrBuildBuildDirEmpty
	}

	if bo.PackageDir == "" {
		return ErrBuildPackageDirEmpty
	}

	if bo.Output == nil {
		bo.Output = io.Discard
	}

	return nil
}

func (co *ConfigOptions) Validate() error {
	// TODO switch + fallthrough
	if co.AlternativeDir == "" {
		return ErrConfigAlternativeDirEmpty
	}

	if co.DatabaseDir == "" {
		return ErrConfigDatabaseDirEmpty
	}

	if co.SourceDir == "" {
		return ErrConfigSourceDirEmpty
	}

	if co.BinaryDir == "" {
		return ErrConfigBinaryDirEmpty
	}

	if co.RootDir == "" {
		co.RootDir = "/"
	}

	st, err := os.Lstat(co.RootDir)

	if err != nil {
		return err
	}

	if !st.IsDir() {
		return ErrConfigRootDirNotDirectory
	}

	return nil
}

func (do *DownloadOptions) Validate() error {
	if do.Progress == nil {
		do.Progress = io.Discard
	}

	return nil
}

func (lo *InstallOptions) Validate() error {
	if lo.ExtractDir == "" {
		return ErrInstallExtractDirEmpty
	}

	return nil
}

func (po *PackageOptions) Validate() error {
	nok := po.Name != ""
	pok := po.Path != ""

	switch {
	case nok && !pok:
		return nil
	case nok && pok:
		return ErrPackagePathNameExclusive
	}

	if filepath.IsAbs(po.Path) {
		return nil
	}

	p, err := filepath.Abs(po.Path)

	if err != nil {
		return err
	}

	po.Path = p
	return nil
}
