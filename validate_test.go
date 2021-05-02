package king_test

import (
	"errors"
	"testing"

	"github.com/illiliti/king"
)

func TestAlternativeOptions(t *testing.T) {
	t.Parallel()

	ss := []struct {
		test     string
		opts     *king.AlternativeOptions
		expected error
	}{
		{
			test: "EmptyPath",
			opts: &king.AlternativeOptions{
				Path: "",
			},
			expected: king.ErrAlternativePathEmpty,
		},
		{
			test: "RelativePath",
			opts: &king.AlternativeOptions{
				Path: "./rel",
			},
			expected: king.ErrAlternativePathNotAbsolute,
		},
		{
			test: "Directory",
			opts: &king.AlternativeOptions{
				Path: t.TempDir(),
			},
			expected: king.ErrAlternativePathDirectory,
		},
		{
			test: "Valid",
			opts: &king.AlternativeOptions{
				Path: "/usr/bin/blkid",
			},
			expected: nil,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			if err := s.opts.Validate(); !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestBuildOptions(t *testing.T) {
	t.Parallel()

	ss := []struct {
		test     string
		opts     *king.BuildOptions
		expected error
	}{
		{
			test: "UnsupportedCompression",
			opts: &king.BuildOptions{
				Compression: "lzip",
			},
			expected: king.ErrBuildCompressionNotSupported,
		},
		{
			test: "EmptyBuildDir",
			opts: &king.BuildOptions{
				BuildDir: "",
			},
			expected: king.ErrBuildBuildDirEmpty,
		},
		{
			test: "EmptyPackageDir",
			opts: &king.BuildOptions{
				BuildDir:   "x",
				PackageDir: "",
			},
			expected: king.ErrBuildPackageDirEmpty,
		},
		{
			test: "Valid",
			opts: &king.BuildOptions{
				PackageDir: "x",
				BuildDir:   "x",
			},
			expected: nil,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			if err := s.opts.Validate(); !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestConfigOptions(t *testing.T) {
	t.Parallel()

	ss := []struct {
		test     string
		opts     *king.ConfigOptions
		expected error
	}{
		{
			test: "EmptyAlternativeDir",
			opts: &king.ConfigOptions{
				AlternativeDir: "",
			},
			expected: king.ErrConfigAlternativeDirEmpty,
		},
		{
			test: "EmptyDatabaseDir",
			opts: &king.ConfigOptions{
				AlternativeDir: "x",
				DatabaseDir:    "",
			},
			expected: king.ErrConfigDatabaseDirEmpty,
		},
		{
			test: "EmptySourceDir",
			opts: &king.ConfigOptions{
				AlternativeDir: "x",
				DatabaseDir:    "x",
				SourceDir:      "",
			},
			expected: king.ErrConfigSourceDirEmpty,
		},
		{
			test: "EmptyBinaryDir",
			opts: &king.ConfigOptions{
				AlternativeDir: "x",
				DatabaseDir:    "x",
				SourceDir:      "x",
				BinaryDir:      "",
			},
			expected: king.ErrConfigBinaryDirEmpty,
		},
		{
			test: "RootDirNotDirectory",
			opts: &king.ConfigOptions{
				AlternativeDir: "x",
				DatabaseDir:    "x",
				SourceDir:      "x",
				BinaryDir:      "x",
				RootDir:        "testdata/TestConfigOptions/regular_file",
			},
			expected: king.ErrConfigRootDirNotDirectory,
		},
		{
			test: "Valid",
			opts: &king.ConfigOptions{
				AlternativeDir: "x",
				DatabaseDir:    "x",
				SourceDir:      "x",
				BinaryDir:      "x",
				RootDir:        "testdata",
			},
			expected: nil,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			if err := s.opts.Validate(); !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestInstallOptions(t *testing.T) {
	t.Parallel()

	ss := []struct {
		test     string
		opts     *king.InstallOptions
		expected error
	}{
		{
			test: "EmptyExtractDir",
			opts: &king.InstallOptions{
				ExtractDir: "",
			},
			expected: king.ErrInstallExtractDirEmpty,
		},
		{
			test: "Valid",
			opts: &king.InstallOptions{
				ExtractDir: "x",
			},
			expected: nil,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			if err := s.opts.Validate(); !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestPackageOptions(t *testing.T) {
	t.Parallel()

	ss := []struct {
		test     string
		opts     *king.PackageOptions
		expected error
	}{
		{
			test: "MutualExclusivity",
			opts: &king.PackageOptions{
				Name: "x",
				Path: "x",
			},
			expected: king.ErrPackagePathNameExclusive,
		},
		{
			test: "Valid",
			opts: &king.PackageOptions{
				Name: "x",
			},
			expected: nil,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			if err := s.opts.Validate(); !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}
