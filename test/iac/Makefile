TEST_STACK := dev

ECA_LINODE_TOKEN := 55772f00-85e0-11f1-85a5-b3f1d391c9f3
export ECA_LINODE_TOKEN

# APL stack test vars
ECA_APL_PROJECT := edgecase-apparel-kubernetes
ECA_APL_STACK := $(TEST_STACK)
export ECA_APL_PROJECT
export ECA_APL_STACK

# Infra stack test vars
ECA_INFRA_PROJECT := edgecase-apparel-infra
ECA_INFRA_STACK := $(TEST_STACK)
export ECA_INFRA_PROJECT
export ECA_INFRA_STACK

go-test:
	go test -v ./unit/...

test-unit: go-test
