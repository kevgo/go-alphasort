# Go-alphasort

[![linux](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml/badge.svg)](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml)
[![windows](https://github.com/kevgo/go-alphasort/actions/workflows/windows.yml/badge.svg)](https://github.com/kevgo/go-alphasort/actions/workflows/windows.yml)

Sorts all elements of Golang source code alphabetically.

- Constructor functions (`New` + type name) and the methods of
  that type move directly behind the type declaration
- All other free-standing functions
  (functions without a receiver)
  get sorted alphabetically within the positions they occupy.
  - Files named `main.go` are scripts, ordered by call flow,
    so their free-standing functions keep their order.
    Everything else in them gets sorted as usual.
- Struct fields get sorted case insensitive.
  Structs containing blank (`_`) fields keep their order,
  since they usually serve as padding.
- Map literals with identifier or string keys get sorted case insensitive by key.
  - Struct instantiations and map literals in
    which more than one value calls a function
    or uses a channel keep their order, since Go evaluates them in source order.
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

## Release

To release a new version, push a tag:

```sh
git sync --all && git checkout main && git tag v0.0.1 && git push --tags
```

Run `make release-snapshot` to try the release build locally.
