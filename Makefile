# Root makefile

-include .env

preview:
	$(MAKE) -C ./cmd preview $(APP)

deploy:
	$(MAKE) -C ./cmd deploy $(APP)

destroy:
	$(MAKE) -C ./cmd destroy $(APP)

test-infra:
	$(MAKE) -C ./test/iac test-unit

go-lint:
	@GOLANGCI_LINT_CACHE=$(mktemp -d) && \
	  golangci-lint fmt && \
	  golangci-lint run

go-fmt:
	@gofmt -l -w ./cmd
	@gofmt -l -w ./test
	@gofumpt -l -w ./cmd ./test
	
test-mate:
	$(MAKE) -C ./test/iac test-var