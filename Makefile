PULUMI_CONFIG_PASSPHRASE = ${ECA_PULUMI_CONFIG_SECRET}
AWS_ACCESS_KEY_ID = ${ECA_OBJ_ACCESS_KEY}
AWS_SECRET_ACCESS_KEY = ${ECA_OBJ_SECRET_KEY}

preview-infra:
	@cd ./cmd/infra && pulumi preview

deploy-infra:
	@cd ./cmd/infra && pulumi up --yes

test-infra:
	@cd ./test/ && go test -v .

go-lint:
	@golangci-lint run

go-fmt:
	@gofumpt -l -w ./cmd ./test
