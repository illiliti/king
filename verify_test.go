package king_test

import (
	"errors"
	"testing"

	"github.com/illiliti/king"
)

func TestVerify(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestVerify"},
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
			test:     "ChecksumMatch",
			expected: nil,
		},
		{
			test:     "ChecksumMismatch",
			expected: king.ErrVerifyMismatch,
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

			vv, err := p.Sources()

			if err != nil {
				t.Fatal(err)
			}

			err = vv[0].(king.Verifier).Verify()

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}
		})
	}
}

func TestSha256(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestSha256"},
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
		result   string
		expected error
	}{
		{
			test:     "RegularFile",
			result:   "c3723545a26246ff2397ba45a6177e32c701df54d7038f14082f9b29f14e4bfb",
			expected: nil,
		},
		{
			test:     "NotRegularFile",
			result:   "",
			expected: king.ErrSha256NotRegular,
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

			vv, err := p.Sources()

			if err != nil {
				t.Fatal(err)
			}

			x, err := vv[0].(king.Verifier).Sha256()

			if !errors.Is(err, s.expected) {
				t.Errorf("got %q, want %q", err, s.expected)
			}

			if x != s.result {
				t.Errorf("got %q, want %q", x, s.result)
			}
		})
	}
}
