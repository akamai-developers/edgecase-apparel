package app

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	cfg "github.com/akamai-developers/edgecase-apparel/cmd/config"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

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

	var (
		apl        cfg.AplConfig
		pulumiConf cfg.PulumiConfig
		stkConf    cfg.PulumiStackConfig
	)

	if err := viper.UnmarshalKey("appPlatform", &apl); err != nil {
		return nil, err
	}

	apl.Token = viper.GetString("linode.token")

	// Get Pulumi stack reference config
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

	// Get OBJ keys and labels from stack reference
	slug := filepath.Join(stkConf.Project, stkConf.Stack)
	stkRef, err := cfg.StackRefInit(ctx, slug)
	if err != nil {
		return nil, err
	}

	objKeys, err := stkRef.Get("obj")
	if err != nil {
		return nil, err
	}

	objBuckets, err := stkRef.Get("objBuckets")
	if err != nil {
		return nil, err
	}

	// Get kubeconfig and create provider
	clusterRefs, err := stkRef.Get("primary")
	if err != nil {
		return nil, err
	}

	kubeconfig, err := decodeKubeconfig(clusterRefs)
	if err != nil {
		return nil, err
	}

	k8sProvider, err := kubernetes.NewProvider(ctx, "ec-kubernetes-provider", &kubernetes.ProviderArgs{
		Kubeconfig: pulumi.String(kubeconfig),
	})
	if err != nil {
		return nil, err
	}

	// Override Helm values
	opts := map[string]any{
		"objKeys":    objKeys,
		"objBuckets": objBuckets,
	}

	values, err := apl.HelmTemplate(opts)
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
	}, pulumi.Provider(k8sProvider), pulumi.DeletedWith(k8sProvider))
	if err != nil {
		return nil, err
	}

	AppPlatformOutputs["apl_id"] = aplChart.ID()
	AppPlatformOutputs["apl_status"] = aplChart.Status
	AppPlatformOutputs["apl_version"] = aplChart.Version

	ctx.Export("apl", AppPlatformOutputs)

	return aplChart, nil
}

func decodeKubeconfig(i any) (string, error) {
	data, ok := i.(map[string]any)
	if !ok {
		err := fmt.Errorf("[ error ] invalid cluster reference type")
		return "", err
	}

	enc, ok := data["kubeconfig"].(string)
	if !ok {
		fmt.Println(data["kubeconfig"].(string))
		err := fmt.Errorf("[ error ] invalid kubeconfig reference type")
		return "", err
	}

	k, err := base64.StdEncoding.DecodeString(string(enc))
	if err != nil {
		return "", err
	}

	return string(k), nil
}
