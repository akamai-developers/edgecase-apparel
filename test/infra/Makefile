TEST_STACK := dev
LOCAL_BACKEND := "file://./state"

ECA_LINODE_TOKEN := 55772f00-85e0-11f1-85a5-b3f1d391c9f3
export ECA_LINODE_TOKEN

# APL stack test vars
ECA_APL_PROJECT := edgecase-apparel-kubernetes
ECA_APL_STACK := $(TEST_STACK)
export ECA_APL_PROJECT ECA_APL_STACK

# Infra stack test vars
ECA_INFRA_PROJECT := edgecase-apparel-infra
ECA_INFRA_STACK := $(TEST_STACK)
export ECA_INFRA_PROJECT
export ECA_INFRA_STACK

clean: pulumi-clean

pulumi-login:
	pulumi login $(LOCAL_BACKEND)

pulumi-init: pulumi-login
	pulumi stack init $(TEST_STACK)
	pulumi stack select $(TEST_STACK)
	pulumi config set linode:token $(ECA_LINODE_TOKEN) --secret

pulumi-setup: pulumi-login
	pulumi stack select $(TEST_STACK)

pulumi-clean: pulumi-login
	pulumi stack rm $(TEST_STACK)

go-test:
	go test -v ./unit/...

#go test -v ./unit/... 

test-unit: go-test
