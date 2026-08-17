package main

import (
	"github.com/akamai-developers/edgecase-apparel/cmd/kubernetes/app"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(app.Deploy)
}
