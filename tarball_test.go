package king_test

import (
	"errors"
	"testing"

	"github.com/illiliti/king"
)

func TestTarballByPath(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   nil,
		AlternativeDir: "x",
		DatabaseDir:    "x",
		BinaryDir:      "x",
		SourceDir:      "x",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	ss := []struct {
		test     string
		path     string
		expected error
	}{
		{
			test:     "Good",
			path:     "testdata/TestTarballByPath/package@1.0-1.tar.gz",
			expected: nil,
		},
		{
			test:     "NoSeparator",
			path:     "testdata/TestTarballByPath/no_separator.tar.gz",
			expected: king.ErrTarballInvalid,
		},
		{
			test:     "NotRegular",
			path:     "testdata/TestTarballByPath/not_regular.tar.gz",
			expected: king.ErrTarballNotRegular,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			_, err := king.NewTarball(c, s.path)

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestTarballByPackage(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestTarballByPackage"},
		AlternativeDir: "x",
		DatabaseDir:    "x",
		BinaryDir:      "testdata/TestTarballByPackage/bin",
		SourceDir:      "x",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	ss := []struct {
		test     string
		expected error
	}{
		{
			test:     "Built",
			expected: nil,
		},
		{
			test:     "NotBuilt",
			expected: king.ErrTarballNotFound,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			p, err := king.NewPackage(c, &king.PackageOptions{
				Name: s.test,
				From: king.Repository,
			})

			if err != nil {
				t.Fatal(err)
			}

			_, err = p.Tarball()

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}
