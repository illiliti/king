package king_test

import (
	"errors"
	"testing"

	"github.com/illiliti/king"
)

func TestSource(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestSource"},
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
		expected error
	}{
		{
			test:     "GoodGit",
			expected: nil,
		},
		{
			test:     "GoodGitWithBranch",
			expected: nil,
		},
		{
			test:     "GoodGitWithCommit",
			expected: nil,
		},
		{
			test:     "GoodHTTP",
			expected: nil,
		},
		{
			test:     "GoodHTTPWithNoExtract",
			expected: nil,
		},
		{
			test:     "GoodRelativeFile",
			expected: nil,
		},
		{
			test:     "GoodAbsoluteFile",
			expected: nil,
		},
		{
			test:     "UnsupportedGitScheme",
			expected: king.ErrSourceGitSchemeInvalid,
		},
		{
			test:     "UnsupportedHTTPScheme",
			expected: king.ErrSourceHTTPSchemeInvalid,
		},
		{
			test:     "RelativeFileNotFound",
			expected: king.ErrSourceFileNotFound,
		},
		{
			test:     "AbsoluteFileNotFound",
			expected: king.ErrSourceFileNotFound,
		},
		{
			test:     "InvalidGitCommit",
			expected: king.ErrSourceGitHashInvalid,
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

			_, err = p.Sources()

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}
