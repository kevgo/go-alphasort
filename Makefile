.DEFAULT_GOAL := help

cuke:  # runs the end-to-end tests (Cucumber scenarios via godog)
	go test -count=1 -run '^TestFeatures$$'

test: unit cuke  # runs all tests

unit:  # runs the unit tests
	go test -count=1 -skip '^TestFeatures$$' ./...

help:  # shows the available targets
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-8s\033[0m %s\n", $$1, $$2}'

.PHONY: cuke test unit help
.SILENT:
