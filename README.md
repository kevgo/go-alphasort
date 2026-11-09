# go-alphasort
Sorts all elements of Golang source code alphabetically.

## Usage

```
go install github.com/kevgo/go-alphasort@latest
go-alphasort [-l] [-w] [path ...]
```

- `-w`: write the sorted result back to the files
- `-l`: list files whose functions aren't sorted
- directories are processed recursively (skipping `vendor`, `testdata`, and hidden directories)
- without paths, reads from STDIN and writes to STDOUT

Free-standing functions (functions without a receiver) get sorted by name in
plain byte order, so capitalized (exported) functions come before lowercase ones. Each function moves together with its doc comment,
directives, and trailing same-line comment. Methods, types, variables,
constants, imports, and floating comments keep their position. The relative
order of functions with identical names (such as multiple `init` functions)
stays intact, since Go executes them in source order.

## Development

```
make test   # all checks and tests (lint, unit, end-to-end)
make lint   # only the static checks (go vet, gofmt, go mod tidy, sorted functions)
make unit   # only the unit tests
make cuke   # only the end-to-end tests (Cucumber scenarios via godog)
make fix    # formats the code and sorts its functions
```
