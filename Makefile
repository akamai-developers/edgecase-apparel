PULUMI_CONFIG_PASSPHRASE = ${ECA_PULUMI_CONFIG_SECRET}
AWS_ACCESS_KEY_ID = ${ECA_OBJ_ACCESS_KEY}
AWS_SECRET_ACCESS_KEY = ${ECA_OBJ_SECRET_KEY}

AUTO_RUN := cd ./cmd/automation && go run main.go

# Requires 's3://[bucket_url]' string exported as $ECA_OBJ_STATE_BACKEND
pulumi-login: pulumi-infra-login pulumi-kube-login

pulumi-kube-login:
	cd ./cmd/kubernetes && \
	  pulumi logout && \
	  pulumi login $(ECA_OBJ_STATE_BACKEND)

pulumi-infra-login:
	cd ./cmd/infra && \
	  pulumi logout && \
	  pulumi login $(ECA_OBJ_STATE_BACKEND)

preview-kube:
	$(AUTO_RUN) preview-kube

deploy-kube:
	$(AUTO_RUN) deploy-kube

destroy-kube:
	$(AUTO_RUN) destroy-kube

preview-infra:
	$(AUTO_RUN) preview-infra

deploy-infra:
	$(AUTO_RUN) deploy-infra

destroy-infra:
	$(AUTO_RUN) destroy-infra

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
	
