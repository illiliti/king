# king
Next generation of the KISS package manager

**Highly experimental, not intended for daily usage. DO NOT RUN ON REAL SYSTEM!**

## Dependencies
* Go >= 1.16 (build time)

## Installation
```sh
# change directory to cli sources
cd cmd/king

# build dynamic binary
go build

# build static binary
go build -tags 'osusergo netgo'

# build static binary and strip debug symbols. best choice!
go build -tags 'osusergo netgo' -ldflags '-s -w'
```

## TODO
* docs
    * [ ] document environment variables
    * [ ] document usage (with examples!!!)
    * [ ] document supported archive formats
    * [ ] document differences between kiss and king
    * [ ] document the whole codebase
        * [x] function/struct description
        * [x] general package documentation
        * [ ] examples
        * [ ] ...
* common
    * [ ] Makefile
    * [ ] unit tests (CI)
    * [ ] improve logging
    * [ ] improve error messages
    * [ ] rework debugging stuff
* future
    * [ ] hooks?
    * [ ] binary packages?
    * [ ] root-less chroot builds?
    * [ ] /var/db/kiss/installed as a git repository?
    * [ ] switch actions to getopt-like flags with subcommands
    * [ ] checksums, sources, version file shouldn't be mandatory
    * [ ] implement privilege elevation mess? is it really needed?
    * [ ] replace list and search actions with query action like in xbps
    * [ ] detach from original KISS but preserve repository layout compatibility?
* action
    * [x] alternative
    * [x] build && update
        * [ ] add a way to skip checksum verification
        * [ ] optionally remove make dependencies after a build
        * [ ] add a way to ignore update for specified package[s]
    * [x] checksum
    * [x] download
    * [x] install
    * [x] list
    * [x] remove
    * [x] search
* library
    * [x] alternative
    * [x] build
        * [x] dynamic dependencies based on ldd-like output
        * [ ] strip binaries using pure go (will be implemented as a standalone project)
        * [x] remove .la and charset.alias
        * [ ] logging to file and stdout
        * [ ] add a way to resume build
        * [ ] handle ctrl+c
    * [x] checksum
    * [ ] config
        * [ ] clean up code
        * [ ] file-based config?
        * [ ] rename essential directories
        * [ ] use system directory for sources/binaries?
    * [x] dependency
        * [ ] guard against circular dependencies
    * [x] download
        * [ ] progress bar
    * [x] install
    * [x] package
    * [x] extract
        * [ ] git clone commit/branch
        * [ ] progress bar for git cloning
    * [x] remove
        * [ ] add a way to forcefully remove files in /etc/
    * [ ] source
        * [x] handle https
        * [ ] handle git
            * [ ] commit
            * [ ] branch
        * [x] handle absolute/relative files
    * [x] tarball
    * [x] update
        * [ ] fix https://github.com/go-git/go-git/issues/37
    * [x] version
* completion
    * [ ] zsh
    * [ ] ...
