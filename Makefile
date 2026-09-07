# Root makefile

-include .env

preview:
	$(MAKE) -C ./cmd preview

deploy:
	$(MAKE) -C ./cmd deploy

destroy:
	$(MAKE) -C ./cmd destroy

test-infra:
	$(MAKE) -C ./test/iac test-unit

go-lint: go-fmt go-lint-main

# enabled linters
go-lint-main:
	@GOLANGCI_LINT_CACHE=$(mktemp -d) && golangci-lint run

# go fomatters
go-fmt:
	@gofmt -l -w ./cmd
	@gofmt -l -w ./test
	@gofumpt -l -w ./cmd ./test
	@golangci-lint fmt
