# Go-alphasort

[![linux](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml/badge.svg)](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml)

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

```sh
go install github.com/kevgo/go-alphasort/cmd/alphasort@latest
```

## Development

Run `make` to see available dev tasks.
