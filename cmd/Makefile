# CMD makefile

-include ../.env
export ECA_KEYCLOAK_ADMIN_PASS
export ECA_KEYCLOAK_CLIENT_SECRET
export ECA_LOKI_ADMIN_PASS
export ECA_CORE_TEAM_PASS
export ECA_SYSDEV_TEAM_PASS
export ECA_GH_TOKEN
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required
export AWS_REQUEST_CHECKSUM_VALIDATION=when_supported

PULUMI_CONFIG_PASSPHRASE = ${ECA_PULUMI_CONFIG_SECRET}
AWS_ACCESS_KEY_ID = ${ECA_OBJ_ACCESS_KEY}
AWS_SECRET_ACCESS_KEY = ${ECA_OBJ_SECRET_KEY}

AUTO_RUN := cd ./automation && go run main.go

# Requires 's3://[bucket_url]' string exported as $ECA_OBJ_STATE_BACKEND
pulumi-login: infra-login kube-login

kube-login:
	cd ./kubernetes && \
	  pulumi logout && \
	  pulumi login $(ECA_OBJ_STATE_BACKEND)

infra-login:
	cd ./infra && \
	  pulumi logout && \
	  pulumi login $(ECA_OBJ_STATE_BACKEND)

rotate-config-passphrase:
	cd ./infra && pulumi stack change-secrets-provider passphrase

kube-linode-token:
	cd ./kubernetes && pulumi config set linode:token $(ECA_LINODE_TOKEN) --secret

infra-linode-token:
	cd ./infra && pulumi config set linode:token $(ECA_LINODE_TOKEN) --secret

rm-stack:
	cd ./kubernetes && pulumi stack rm dev --force -y

rotate-linode-token: infra-linode-token kube-linode-token 

preview:
	$(AUTO_RUN) "preview" $(APP)

deploy:
	$(AUTO_RUN) "deploy" $(APP)

destroy:
	$(AUTO_RUN) "destroy" $(APP)

cancel:
	cd ./kubernetes && pulumi cancel