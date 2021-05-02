package king_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/illiliti/king"
)

// TODO test package with multiple sources that may overlap
// TODO test copy empty directories
func TestExtract(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestExtract"},
		AlternativeDir: "x",
		DatabaseDir:    "x",
		BinaryDir:      "x",
		SourceDir:      "testdata/TestExtract/src",
		RootDir:        "testdata",
	})

	if err != nil {
		t.Fatal(err)
	}

	ss := []struct {
		test  string
		check string
	}{
		// TODO fix nested directories inside files/ or ban them
		// NOTE currently we match behavior with kiss and i don't think that this is needing "fixing"
		// {
		// 	test:  "CopyDirectory",
		// 	check: "directory/data",
		// },
		// {
		// 	test:  "CopyDirectoryWithExtractDir",
		// 	check: "directory/data",
		// },
		{
			test:  "CopyFile",
			check: "text_file",
		},
		{
			test:  "CopyFileWithExtractDir",
			check: "text_file",
		},
		{
			test:  "CopyHTTP",
			check: "src.tar.gz",
		},
		{
			test:  "CopyHTTPWithExtractDir",
			check: "src.tar.gz",
		},
		{
			test:  "ExtractHTTP",
			check: "regular_file",
		},
		{
			test:  "ExtractHTTPWithExtractDir",
			check: "regular_file",
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

			ss, err := p.Sources()

			if err != nil {
				t.Fatal(err)
			}

			d := filepath.Join(t.TempDir(), ss[0].ExtractDir())

			err = ss[0].Extract(d)

			if err != nil {
				t.Fatal(err)
			}

			_, err = os.Stat(filepath.Join(d, s.check))

			if err != nil {
				t.Errorf("Extract() incorrectly unpacked %q into %q", s.check, d)
			}
		})
	}
}

func TestExtractDir(t *testing.T) {
	t.Parallel()

	c, err := king.NewConfig(&king.ConfigOptions{
		Repositories:   []string{"testdata/TestExtractDir"},
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
		test   string
		result string
	}{
		{
			test:   "GoodExtractDir",
			result: "subdir",
		},
		{
			test:   "NoExtractDir",
			result: "",
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

			ss, err := p.Sources()

			if err != nil {
				t.Fatal(err)
			}

			d := ss[0].ExtractDir()

			if d != s.result {
				t.Errorf("got %q, want %q", d, s.result)
			}
		})
	}
}
