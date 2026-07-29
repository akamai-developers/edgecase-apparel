package app

import (
	"fmt"
	"path/filepath"
	"strings"

	cfg "github.com/akamai-developers/edgecase-apparel/cmd/config"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

// const infaProject string = "organization/edgecase-apparel-infra"

var AppPlatformOutputs = pulumi.Map{}

func Deploy(ctx *pulumi.Context) error {
	if _, err := DeployApl(ctx); err != nil {
		return err
	}

	return nil
}

func DeployApl(ctx *pulumi.Context) (*helm.Release, error) {
	// Get Viper configs
	if err := cfg.InitConfig(); err != nil {
		return nil, err
	}

	// Get Pulumi stack reference config
	var (
		pulumiConf cfg.PulumiConfig
		stkConf    cfg.PulumiStackConfig
	)

	if err := viper.UnmarshalKey("pulumi", &pulumiConf); err != nil {
		return nil, err
	}

	stkContext := pulumiConf.Context
	for _, i := range pulumiConf.Projects {
		prefix, suffix, _ := strings.Cut(i.Name, "-")

		if prefix == "infra" && suffix == stkContext[prefix] {
			stkConf = i
		}
	}

	// Get OBJ key pair from stack reference
	slug := filepath.Join(stkConf.Project, stkConf.Stack)
	// st, err := pulumi.NewStackReference(ctx, slug, nil)

	// out, err := st.GetOutputDetails("obj")

	// outputs, err := pulumi.NewStackReference("dns")

	stkRef, err := cfg.StackRefInit(ctx, slug)
	if err != nil {
		return nil, err
	}

	objKeys, err := stkRef.GetMap("obj")
	if err != nil {
		return nil, err
	}

	// Get OBJ bucket lables
	objBuckets, err := stkRef.GetMap("objBuckets")
	fmt.Println(objBuckets)
	if err != nil {
		return nil, err
	}

	var apl cfg.AplConfig
	if err := viper.UnmarshalKey("appPlatform", &apl); err != nil {
		return nil, err
	}

	// Get Linode API token
	apl.Token = viper.GetString("linode.token")

	// Override Helm values
	values, err := apl.HelmTemplate(objKeys, objBuckets)
	if err != nil {
		return nil, err
	}

	// Deploy APL Helm chart
	aplChart, err := helm.NewRelease(ctx, apl.Name, &helm.ReleaseArgs{
		Chart:           pulumi.String(apl.Chart),
		CreateNamespace: pulumi.Bool(true),
		DisableWebhooks: pulumi.Bool(false),
		Lint:            pulumi.Bool(true),
		Name:            pulumi.String(apl.Chart),
		RepositoryOpts: helm.RepositoryOptsArgs{
			Repo: pulumi.String(apl.Repo),
		},
		ReuseValues: pulumi.Bool(true),
		Timeout:     pulumi.Int(1200),
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewStringAsset(values),
		},
		WaitForJobs: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	AppPlatformOutputs["apl_id"] = aplChart.ID()
	AppPlatformOutputs["apl_status"] = aplChart.Status
	AppPlatformOutputs["apl_version"] = aplChart.Version

	ctx.Export("apl", AppPlatformOutputs)

	return aplChart, nil
}
