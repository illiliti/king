package king_test

import (
	"errors"
	"testing"

	"github.com/illiliti/king"
)

func TestPackageByName(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestPackageByName"},
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
		from     king.RepositoryType
		expected error
	}{
		{
			test:     "FromAll",
			from:     king.All,
			expected: nil,
		},
		{
			test:     "FromRepository",
			from:     king.Repository,
			expected: nil,
		},
		{
			test:     "FromDatabase",
			from:     king.Database,
			expected: king.ErrPackageNameNotFound,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			_, err := king.NewPackage(c, &king.PackageOptions{
				Name: s.test,
				From: s.from,
			})

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestPackageByPath(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   nil,
		AlternativeDir: "x",
		DatabaseDir:    "TestPackageByPath",
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
			test:     "NotOwnedFile",
			path:     "/some/nonexistent/path",
			expected: king.ErrPackagePathNotFound,
		},
		{
			test:     "NotOwnedDir",
			path:     "/some/nonexistent/dir/",
			expected: king.ErrPackagePathNotFound,
		},
		{
			test:     "OwnedFile",
			path:     "/some/random/path",
			expected: nil,
		},
		{
			test:     "OwnerDir",
			path:     "/some/random/dir/",
			expected: king.ErrPackagePathNotFound, // XXX error is misleading
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			_, err := king.NewPackage(c, &king.PackageOptions{
				Path: s.path,
			})

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestMultiplePackagesOwnsSamePath(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   nil,
		AlternativeDir: "x",
		DatabaseDir:    "TestMultiplePackagesOwnsSamePath",
		BinaryDir:      "x",
		SourceDir:      "x",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		recover()
	}()

	king.NewPackage(c, &king.PackageOptions{
		Path: "/some/path",
	})

	t.Error("unreachable")
}
