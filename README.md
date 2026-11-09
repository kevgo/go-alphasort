# go-alphasort

Sorts all elements of Golang source code alphabetically.

## Usage

```text
go install github.com/kevgo/go-alphasort/cmd/alphasort@latest
alphasort fix          # sort the functions in all Go files in the current directory and its subdirectories
alphasort fix <files>  # sort the functions in the given files only
alphasort help         # show help
```

`fix` rewrites the files on disk. Without arguments, it skips `vendor`,
`testdata`, hidden, and `_`-prefixed directories.

Alphasort reorders the functions and methods in each file:

- The constructor of a type (the function named `New` + type name, e.g.
  `NewCircle` for `Circle`) and the methods of that type move directly behind
  the type declaration: first the constructor, then the methods sorted
  alphabetically.
- All other free-standing functions (functions without a receiver) get sorted
  alphabetically within the positions they occupy.

Names compare in plain byte order, so capitalized (exported) names come before
lowercase ones. Each function and method moves together with its doc comment,
directives, and trailing same-line comment. Variables, constants, imports,
floating comments, and methods of types declared in other files keep their
position. The relative order of functions with identical names (such as
multiple `init` functions) stays intact, since Go executes them in source
order.

## Development

```text
make test   # all checks and tests (lint, unit, end-to-end)
make lint   # only the static checks (go vet, go mod tidy)
make unit   # only the unit tests
make cuke   # only the end-to-end tests (Cucumber scenarios via godog)
make fix    # formats the code and sorts its functions
```
