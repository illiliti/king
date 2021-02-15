module github.com/illiliti/king

go 1.15

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/go-git/go-git/v5 v5.2.0
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/henvic/ctxsignal v1.0.0
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/klauspost/compress v1.11.7 // indirect
	github.com/mholt/archiver/v3 v3.5.0
	github.com/pierrec/lz4/v4 v4.1.3 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

// see https://github.com/mholt/archiver/pull/265
replace github.com/mholt/archiver/v3 => github.com/illiliti/archiver/v3 v3.3.2-0.20210214122530-1f8df2101bad
