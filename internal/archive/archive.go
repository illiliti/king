package archive

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
)

// TODO unit tests

func Create(s, d string) error {
	f, err := archiver.ByExtension(d)

	if err != nil {
		return err
	}

	switch v := f.(type) {
	case *archiver.Zip:
		v.OverwriteExisting = true
	case *archiver.Tar:
		v.OverwriteExisting = true
	case *archiver.TarBrotli:
		v.OverwriteExisting = true
	case *archiver.TarBz2:
		v.OverwriteExisting = true
	case *archiver.TarGz:
		v.OverwriteExisting = true
	case *archiver.TarLz4:
		v.OverwriteExisting = true
	case *archiver.TarSz:
		v.OverwriteExisting = true
	case *archiver.TarXz:
		v.OverwriteExisting = true
	case *archiver.TarZstd:
		v.OverwriteExisting = true
	default:
		return fmt.Errorf("create archive %s: unsupported format", s)
	}

	if err := os.MkdirAll(filepath.Dir(d), 0777); err != nil {
		return err
	}

	wd, err := os.Getwd()

	if err != nil {
		return err
	}

	if err := os.Chdir(s); err != nil {
		return err
	}

	defer os.Chdir(wd)
	return f.(archiver.Archiver).Archive([]string{"."}, d)
}

func Extract(s, d string, c int) error {
	f, err := archiver.ByExtension(s)

	if err != nil {
		return err
	}

	switch v := f.(type) {
	case *archiver.Zip:
		v.StripComponents = c
	case *archiver.Rar:
		v.StripComponents = c
	case *archiver.Tar:
		v.StripComponents = c
	case *archiver.TarBrotli:
		v.Tar.StripComponents = c
	case *archiver.TarBz2:
		v.Tar.StripComponents = c
	case *archiver.TarGz:
		v.Tar.StripComponents = c
	case *archiver.TarLz4:
		v.Tar.StripComponents = c
	case *archiver.TarSz:
		v.Tar.StripComponents = c
	case *archiver.TarXz:
		v.Tar.StripComponents = c
	case *archiver.TarZstd:
		v.Tar.StripComponents = c
	default:
		return fmt.Errorf("extract archive %s: unsupported format", s)
	}

	return f.(archiver.Unarchiver).Unarchive(s, d)
}
