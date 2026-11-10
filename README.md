# Go-alphasort

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

```text
make test   # all checks and tests (lint, unit, end-to-end)
make lint   # only the static checks (go vet, go mod tidy)
make unit   # only the unit tests
make cuke   # only the end-to-end tests (Cucumber scenarios via godog)
make fix    # formats the code and sorts its functions
```
