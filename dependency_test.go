package king_test

import (
	"testing"

	"github.com/illiliti/king"
)

func TestDependencies(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata"},
		AlternativeDir: "x",
		DatabaseDir:    "x",
		BinaryDir:      "x",
		SourceDir:      "x",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: "TestDependencies",
		From: king.Repository,
	})

	if err != nil {
		t.Fatal(err)
	}

	ss := []struct {
		test     string
		result   *king.Dependency
		expected bool
	}{
		{
			test: "Match",
			result: &king.Dependency{
				Name:   "some_dependency",
				IsMake: false,
			},
			expected: true,
		},
		{
			test: "Mismatch",
			result: &king.Dependency{
				Name:   "some_nonexistent_dependency",
				IsMake: false,
			},
			expected: false,
		},
		{
			test: "MismatchMake",
			result: &king.Dependency{
				Name:   "some_dependency",
				IsMake: true,
			},
			expected: false,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			dd, err := p.Dependencies()

			if err != nil {
				t.Fatal(err)
			}

			if (*dd[0] == *s.result) != s.expected {
				t.Errorf("got %q, want %q", dd[0], s.result)
			}
		})
	}
}

func TestDependsOnItself(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata"},
		AlternativeDir: "x",
		DatabaseDir:    "x",
		BinaryDir:      "x",
		SourceDir:      "x",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: "TestDependsOnItself",
		From: king.Repository,
	})

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		recover()
	}()

	p.Dependencies()
	t.Error("unreachable")
}

func TestCircularDependencies(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestCircularDependencies"},
		AlternativeDir: "x",
		DatabaseDir:    "x",
		BinaryDir:      "x",
		SourceDir:      "x",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: "package1",
		From: king.Repository,
	})

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		recover()
	}()

	p.RecursiveDependencies()
	t.Error("unreachable")
}

func TestReverseDependencies(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   nil,
		AlternativeDir: "x",
		DatabaseDir:    "TestReverseDependencies",
		BinaryDir:      "x",
		SourceDir:      "x",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	p, err := king.NewPackage(c, &king.PackageOptions{
		Name: "package1",
		From: king.Database,
	})

	if err != nil {
		t.Fatal(err)
	}

	ss := []struct {
		test     string
		result   string
		expected bool
	}{
		{
			test:     "Match",
			result:   "package2",
			expected: true,
		},
		{
			test:     "Mismatch",
			result:   "package3",
			expected: false,
		},
	}

	for _, s := range ss {
		s := s // HACK

		t.Run(s.test, func(t *testing.T) {
			t.Parallel()

			dd, err := p.ReverseDependencies()

			if err != nil {
				t.Fatal(err)
			}

			if (dd[0] == s.result) != s.expected {
				t.Errorf("got %q, want %q", dd[0], s.result)
			}
		})
	}
}
