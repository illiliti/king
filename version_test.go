package king_test

import (
	"errors"
	"testing"

	"github.com/illiliti/king"
)

// TODO better docs
func TestVersion(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestVersion"},
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
			test:     "WithoutRelease",
			expected: king.ErrVersionInvalid,
		},
		{
			test:     "WithRelease",
			expected: nil,
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

			_, err = p.Version()

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}
