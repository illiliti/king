package hash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// TODO unit tests

func Random(c int) string {
	cr := make([]byte, c)

	if _, err := rand.Read(cr); err != nil {
		panic(err)
	}

	return hex.EncodeToString(cr)
}

func Sha256(p string) (string, error) {
	f, err := os.Open(p)

	if err != nil {
		return "", err
	}

	defer f.Close()

	c := sha256.New()

	if _, err := io.Copy(c, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(c.Sum(nil)), nil
}
