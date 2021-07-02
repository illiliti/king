package main

// TODO move examples to man page

const (
	kingUsage = `usage: king [option]... <action> [<package>]...

option:
    --repository <dir>  set directory to repository of packages
    --binary-dir <dir>  set directory to pre-built packages
    --source-dir <dir>  set directory to package sources
    --root-dir   <dir>  set directory to root
    -v, --version       print current version
    -h, --help          print help message

action:
    (b)uild     build package
    (c)hecksum  generate checksums
    (d)ownload  download sources
    (i)nstall   install package
    (q)uery     query metadata
    (r)emove    remove package
    (s)wap      swap conflict
    (u)pdate    update system

example:
    king --root-dir /mnt/bootstrap_rootfs i baselayout busybox
    king --repository /mnt/repo/core,/mnt/repo/extra build firefox`

	queryUsage = `usage: king query [option]... [<package>]...

option:
    -R, --only-repositories  query data only from repositories
    -B, --only-database      query data only from database
    -l, --packages           list installed packages
    -L, --repositories       list available repositories
    -A, --alternatives       list occurred conflicts
    -s, --search             search package across repositories
    -1, --single             output only first match during search operation
    -z, --size       <pkg>   display size of package files
    -f, --manifest   <pkg>   display files owned by package
    -m, --maintainer <pkg>   display maintainer contacts of package
    -d, --deps       <pkg>   display package dependencies
    -D, --revdeps    <pkg>   display which packages depend on given package
    -o, --owner      <pkg>   display which package owns given path
    -h, --help               print help message

example:
    king q -s -1 dwm slstatus
    king query -Rd firefox`

	// -r, --remove-make        remove make dependencies after all builds are complete
	// -O, --output-dir  <dir>  set directory where build logs will be written
	// -S, --no-strip           disable binary stripping
	// TODO add option to always prompt
	buildUsage = `usage: king build [option]... <package>...

option:
    -X, --extract-dir <dir>  set directory where pre-built packages will be extracted
    -P, --package-dir <dir>  set directory where package will be turned into tarball
    -B, --build-dir   <dir>  set directory where sources will be extracted
    -C, --compression <fmt>  select compression format
    -s, --no-verify          skip checksum verification for sources
    -d, --debug              keep build, package, extract directory after build is complete
    -f, --force              re-download sources even if they are already exist
    -n, --no-bar             disable progress bar for downloadable sources
    -y, --no-prompt          disable confirmation prompt
    -T, --no-prebuilt        rebuild pre-built packages
    -i, --install            install built packages
    -q, --quiet              suppress build script output
    -h, --help               print help message

example:
    king build -C zst -q --install mesa
    king b -X /tmp/extract -P /tmp/package -B /tmp/build llvm clang`

	// -r, --remove-make        remove make dependencies after all updates are complete
	// -O, --output-dir  <dir>  set directory where build logs will be written
	updateUsage = `usage: king update [option]...

option:
    -X, --extract-dir <dir>  set directory where pre-built packages will be extracted
    -P, --package-dir <dir>  set directory where packages will be turned into tarballs
    -B, --build-dir   <dir>  set directory where sources will be extracted
    -C, --compression <fmt>  select compression format
    -x, --exclude     <pkg>  ignore update for specified package
    -s, --no-verify          skip checksum verification for sources
    -d, --debug              keep build, package, extract directory after update is complete
    -f, --force              re-download sources even if they are already exist
    -n, --no-bar             disable progress bar for downloadable sources
    -y, --no-prompt          disable confirmation prompt
    -N, --no-pull            disable updating repositories
    -c, --no-error           continue on error during updating repositories
    -T, --no-prebuilt        rebuild pre-built packages
    -q, --quiet              suppress build script output
    -h, --help               print help message

example:
    king u -fq
    king update --no-pull -e firefox,llvm,rust,cmake`

	installUsage = `usage: king install [option]... <package>...

option:
    -X, --extract-dir <dir>  set directory where binary packages will be extracted
    -d, --debug              keep extract directory after install is complete
    -f, --force              install package even if dependencies are unmet
    -e, --overwrite-etc      overwrite /etc/* files without special handling
    -h, --help               print help message

example:
    king i libudev-zero
    king install -X /tmp/extract -e sway`

	removeUsage = `usage: king remove [option]... <package>...

option:
    -f, --force       remove package even if other packages depend on it
    -e, --remove-etc  remove /etc/* files without special handling
    -r, --recursive   recursively remove dependencies
    -h, --help        print help message

example:
    king r -e eudev
    king remove -rf xorg-server`

	downloadUsage = `usage: king download [option]... <package>...

option:
    -f, --force   re-download sources even if they are already exist
    -n, --no-bar  disable progress bar
    -h, --help    print help message

example:
    king d -n dash
    king download -f firefox`

	swapUsage = `usage: king swap [option]... <path>...
       king swap [-]

option:
    -t, --target <pkg>  swap implementation to given package
    -h, --help          print help message

example:
    king query --alternatives util-linux | king swap
    king s -t util-linux /usr/bin/mount /usr/bin/blkid
    king swap /usr/bin/patch /usr/bin/find /usr/bin/grep`

	checksumUsage = `usage: king checksum [option]... <package>...

option:
    -h, --help  print help message

example:
    king checksum firefox
    king c gcc clang king`
)
