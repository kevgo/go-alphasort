# Go-alphasort

[![linux](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml/badge.svg)](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml)
[![windows](https://github.com/kevgo/go-alphasort/actions/workflows/windows.yml/badge.svg)](https://github.com/kevgo/go-alphasort/actions/workflows/windows.yml)

Sorts all elements of Golang source code alphabetically.

- Constructor functions (`New` + type name) and the methods of
  that type move directly behind the type declaration
- All other free-standing functions
  (functions without a receiver)
  get sorted alphabetically within the positions they occupy.
- Skips `vendor`, `testdata`, hidden, and `_`-prefixed directories.
- The relative order of functions with identical names
  (such as multiple `init` functions)
  stays intact, since Go executes them in source order.

## Usage

```sh
alphasort fix          # format all files
alphasort fix <files>  # format the given files
```

## Installation

Download a prebuilt binary from the
[releases page](https://github.com/kevgo/go-alphasort/releases) or install from
source:

```sh
go install github.com/kevgo/go-alphasort/cmd/alphasort@latest
```

## Development

Run `make` to see available dev tasks.

To release a new version, push a tag:

```sh
git tag v1.2.3
git push origin v1.2.3
```

This triggers the release workflow,
which uses [GoReleaser](https://goreleaser.com) to build binaries for Linux,
macOS, and Windows and publish them to GitHub Releases.
Run `make release-snapshot` to try the release build locally.
