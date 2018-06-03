# go-squish
[![GoDoc](https://godoc.org/github.com/InfinityTools/go-squish?status.svg)](https://godoc.org/github.com/InfinityTools/go-squish)

## About

[libsquish](https://sourceforge.net/projects/libsquish/) is a small, portable C++ library for compressing and decompressing images with the DXT standard (also known as S3TC or BC). This standard is mainly used by OpenGL and DirectX for the lossy compression of RGBA textures..

This project provides [Go](https://golang.org/) bindings for *libsquish*. *go-squish* also provides a patch file that adds a C compatibility wrapper to libsquish, since Go doesn't support C++ libraries directly. Apply it before compiling the library. It is compatible with libsquish version 1.15.

## Building

go-squish uses the system version of the static *libisquish* library. More information about how to build libimagequant can be found in the [libisquish readme](https://sourceforge.net/projects/libsquish/files/).

go-squish package path is currently `github.com/InfinityTools/squish`. The bindings can be built via `go build`.

This package makes use of CGO, which requires a decent C compiler to be installed. However, using `go install` removes the C compiler requirement for future invocations of `go build`.

## Documentation

For docs, see https://godoc.org/github.com/InfinityTools/go-squish .

## License

Both *go-squish* and the libsquish C wrapper patch are released under the BSD 2-clause license. See LICENSE for more details.

*libsquish* itself is available under under the *MIT license*. See *libsquish-license.txt* for more details.
