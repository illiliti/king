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
