# king
Next generation of the KISS package manager

**Unstable, not intended for daily usage**

## Dependencies
* Go >= 1.15 (build time)

## Installation
```sh
# change directory
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
    * [ ] document the whole codebase (with examples!!!)
* common
    * [ ] Makefile
    * [ ] unit tests (CI)
    * [ ] improve logging
    * [ ] improve error messages
* future
    * [ ] hooks?
    * [ ] binary packages?
    * [ ] root-less chroot builds?
    * [ ] switch actions to getopt-like flags
    * [ ] implement privilege elevation mess? is it really needed?
    * [ ] replace list and search actions with query action like in xbps
    * [ ] detach from original KISS but preserve repository layout compatibility?
* action
    * [x] alternative
    * [ ] build
        * [ ] rewrite logging messages
        * [ ] ask confirmation about building
        * [ ] add a way to skip checksum verification
    * [ ] checksum
        * [ ] fix empty checksums file if sources file only contains git source
        * [ ] rewrite logging messages
    * [x] download
    * [x] install
    * [x] list
    * [x] remove
    * [x] search
    * [ ] update
* library
    * [x] alternative
    * [ ] build
        * [ ] dynamic dependencies based on ldd-like output
        * [ ] strip binaries using pure go (will be implemented as separate project)
        * [ ] remove .la and charset.alias
        * [ ] logging to file and stdout
        * [ ] add a way to resume build
        * [ ] handle ctrl+c
    * [x] checksum
        * [ ] solve inconsistent mess between checksum.go and internal/chksum/chksum.go
    * [ ] config
        * [ ] file-based config?
        * [ ] rename essential directories
        * [ ] use system directory for sources/binaries?
    * [ ] dependency
        * [ ] guard against cyclic dependencies
    * [ ] download
        * [ ] progress bar
        * [ ] drop ctxsignal
    * [ ] install
        * [ ] guard against incomplete installation
        * [ ] ensure that /etc/ handling is working correctly
    * [x] package
    * [ ] prepare
        * [ ] ensure git support
    * [ ] remove
        * [ ] guard against incomplete removal
        * [ ] automagically swap dangling alternatives?
    * [ ] source
        * [x] handle https
        * [ ] handle git (need to recheck)
        * [x] handle absolute/relative files
    * [x] tarball
    * [ ] update
        * [ ] builtin "kiss-outdated"
        * [ ] speed up via concurrency
    * [x] version
