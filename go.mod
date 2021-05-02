module github.com/illiliti/king

go 1.16

require (
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/cornfeedhobo/pflag v1.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/go-git/go-billy/v5 v5.3.0 // indirect
	github.com/go-git/go-git/v5 v5.3.0
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/klauspost/compress v1.12.2 // indirect
	github.com/mholt/archiver/v3 v3.5.0
	github.com/pierrec/lz4/v4 v4.1.6 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/net v0.0.0-20210502030024-e5908800b52b // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887
)

// XXX https://github.com/mholt/archiver/pull/265
replace github.com/mholt/archiver/v3 => github.com/illiliti/archiver/v3 v3.3.2-0.20210214122530-1f8df2101bad
