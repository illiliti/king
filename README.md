# king
Next generation of the KISS package manager

**Unstable but usable. Breakage is still possible though**

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

# build static binary, strip debug symbols and trim internal paths. best choice!
go build -tags 'osusergo netgo' -ldflags '-s -w' -trimpath
```
