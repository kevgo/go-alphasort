# Go-alphasort

[![linux](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml/badge.svg)](https://github.com/kevgo/go-alphasort/actions/workflows/linux.yml)
[![windows](https://github.com/kevgo/go-alphasort/actions/workflows/windows.yml/badge.svg)](https://github.com/kevgo/go-alphasort/actions/workflows/windows.yml)

Sorts all elements of Golang source code alphabetically.

- Constructor functions (`New` + type name) and the methods of
  that type move directly behind the type declaration
- All other free-standing functions
  (functions without a receiver)
  get sorted alphabetically within the positions they occupy.
- Struct fields get sorted alphabetically.
  Structs containing blank (`_`) fields keep their order,
  since they usually serve as padding.
  - Instantiations in which more than one field value calls a function
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

```sh
go install github.com/kevgo/go-alphasort/cmd/alphasort@latest
```

## Development

Run `make` to see available dev tasks.
