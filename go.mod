module github.com/illiliti/king

go 1.16

require (
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20210512092938-c05353c2d58c // indirect
	github.com/andybalholm/brotli v1.0.3 // indirect
	github.com/cornfeedhobo/pflag v1.1.0
	github.com/dustin/go-humanize v1.0.0
	github.com/go-git/go-git/v5 v5.4.2
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/mholt/archiver/v3 v3.5.0
	github.com/pierrec/lz4/v4 v4.1.8 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
)

// XXX https://github.com/mholt/archiver/pull/265
replace github.com/mholt/archiver/v3 => github.com/illiliti/archiver/v3 v3.3.2-0.20210214122530-1f8df2101bad
