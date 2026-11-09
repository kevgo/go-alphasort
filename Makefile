RTA_VERSION = 0.40.0

RTA          = tools/rta@$(RTA_VERSION)
ACTIONLINT   = $(RTA) actionlint
CUCUMBERSORT = $(RTA) cucumber-sort
DEADCODE     = $(RTA) deadcode
GHOKIN       = $(RTA) ghokin
GOFUMPT      = $(RTA) gofumpt
GOLANGCILINT = $(RTA) golangci-lint
RUMDL        = $(RTA) rumdl
STATICCHECK  = $(RTA) --from-source staticcheck
TAPLO        = $(RTA) taplo

cuke: install  # runs the end-to-end tests
	go test -count=1 -run '^TestFeatures$$'

fix: install  # formats the code
	~/go/bin/alphasort fix
	$(GOFUMPT) -l -w .
	$(RUMDL) fmt
	$(TAPLO) format
	$(CUCUMBERSORT) format
	$(GHOKIN) fmt replace features/

help:  # shows the available targets
	grep -h -E '^[a-zA-Z0-9_-]+:.*?# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-8s\033[0m %s\n", $$1, $$2}'

install:  # builds for the current platform
	go install -ldflags="-s -w" ./cmd/alphasort

lint:  # runs the static checks
	go vet ./...
	$(DEADCODE) -test github.com/kevgo/go-alphasort
	$(ACTIONLINT)
	$(STATICCHECK) ./...
	$(GOLANGCILINT) cache clean && $(GOLANGCILINT) run

test: lint unit cuke  # runs all checks and tests

unit:  # runs the unit tests
	go test -count=1 -skip '^TestFeatures$$' ./...

update:
	go mod tidy


.PHONY: cuke fix help install lint test unit update
.SILENT:
.DEFAULT_GOAL := help
