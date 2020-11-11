package main

import (
	"fmt"
	"os"

	"github.com/illiliti/king"
	"github.com/illiliti/king/internal/log"
)

const usage = `usage: king [a|b|c|d|i|l|r|s|u|v] [package ...]

alternative List or swap to alternatives
build       Build a package
checksum    Generate checksums
download    Pre-download all sources
install     Install a package
list        List installed packages
remove      Remove a package
search      Search for a package
update      Update installed packages
version     Show package manager version`

// TODO refactor 'not enough arguments'

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	c, err := king.NewContext()

	if err != nil {
		log.Fatal(err)
	}

	switch args := os.Args[2:]; os.Args[1] {
	case "alternative", "a":
		alternative(c, args)
	case "build", "b":
		build(c, args)
	case "checksum", "c":
		checksum(c, args)
	case "download", "d":
		download(c, args)
	case "install", "i":
		install(c, args)
	case "list", "l":
		list(c, args)
	case "remove", "r":
		remove(c, args)
	case "search", "s":
		search(c, args)
	case "update", "u":
		update(c)
	case "version", "v":
		log.Info("PRE-ALPHA")
	default:
		log.Fatal("invalid action:", os.Args[1])
	}
}
