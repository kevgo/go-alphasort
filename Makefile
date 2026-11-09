.DEFAULT_GOAL := help

cuke:  # runs the end-to-end tests (Cucumber scenarios via godog)
	go test -count=1 -run '^TestFeatures$$'

fix:  # formats the code and sorts its functions
	go mod tidy
	go run ./cmd/alphasort fix .
	gofmt -w .

lint:  # runs the static checks
	go vet ./...
	go mod tidy -diff
	files=$$(go run ./cmd/alphasort check .); if [ -n "$$files" ]; then echo "functions not sorted (run \"make fix\"):"; echo "$$files"; exit 1; fi

test: lint unit cuke  # runs all checks and tests

unit:  # runs the unit tests
	go test -count=1 -skip '^TestFeatures$$' ./...

help:  # shows the available targets
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-8s\033[0m %s\n", $$1, $$2}'

.PHONY: cuke fix help lint test unit
.SILENT:
