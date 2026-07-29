PULUMI_CONFIG_PASSPHRASE = ${ECA_PULUMI_CONFIG_SECRET}
AWS_ACCESS_KEY_ID = ${ECA_OBJ_ACCESS_KEY}
AWS_SECRET_ACCESS_KEY = ${ECA_OBJ_SECRET_KEY}

preview-infra-old:
	@cd ./cmd/infra && pulumi preview

preview-infra:
	@cd ./cmd/automation && go run main.go preview-infra

deploy-infra:
	@cd ./cmd/automation && go run main.go deploy-infra

test-infra:
	cd ./test/infra && make test-unit

go-lint:
	@golangci-lint run

go-fmt:
	@gofumpt -l -w ./cmd ./test
