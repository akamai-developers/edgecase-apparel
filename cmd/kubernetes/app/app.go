package app

import (
	"encoding/base64"
	"errors"

	cfg "github.com/akamai-developers/edgecase-apparel/cmd/config"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

var (
	// AppPlatformOutputs is a Pulumi map for exporting App Platform outputs.
	AppPlatformOutputs  = pulumi.Map{}
	KubeProviderOutputs = pulumi.Map{}
	errKubeConfig       = errors.New("[ error ] invalid kubeconfig reference type")
)

// Deploy is a wrapper around the DeployAPL func, which enables the Pulumi CLI
// to execute the program independently of the Automation API.
func Deploy(ctx *pulumi.Context) error {
	if _, err := DeployApl(ctx); err != nil {
		return err
	}

	return nil
}

// DeployApl is a Pulumi program that deploys the App Platform Helm chart to a
// Kubernetes cluster.
//
//nolint:funlen
func DeployApl(ctx *pulumi.Context) (*helm.Release, error) {
	// Get Viper configs
	if err := cfg.InitConfig(); err != nil {
		return nil, err
	}

	var apl cfg.AplConfig

	if err := viper.UnmarshalKey("appPlatform", &apl); err != nil {
		return nil, err
	}

	apl.Token = viper.GetString("linode.token")

	// Initialize infra stack references
	stkRef, err := cfg.StackRefInit(ctx, "infra")
	if err != nil {
		return nil, err
	}

	// Get kubeconfig, decode it, and create provider
	kubecfg, err := stkRef.GetKubeConfig("primary")
	if err != nil {
		return nil, err
	}

	kubeconfig, err := decodeKubeconfig(kubecfg)
	if err != nil {
		return nil, err
	}

	k8sProvider, err := kubernetes.NewProvider(ctx, "ec-kubernetes-provider", &kubernetes.ProviderArgs{
		Kubeconfig: pulumi.String(kubeconfig),
	})
	if err != nil {
		return nil, err
	}

	AppPlatformOutputs["k8sProviderId"] = k8sProvider.ID()

	// Get OBJ keys and bucket labels
	objData, err := stkRef.GetObj()
	if err != nil {
		return nil, err
	}

	// Override Helm values template variables
	values, err := apl.HelmTemplate(objData)
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

	// Export APL stack outputs
	AppPlatformOutputs["aplId"] = aplChart.ID()
	AppPlatformOutputs["aplLint"] = aplChart.Lint
	AppPlatformOutputs["aplName"] = aplChart.Name
	AppPlatformOutputs["aplRepo"] = aplChart.RepositoryOpts.Repo()
	AppPlatformOutputs["aplStatus"] = aplChart.Status
	AppPlatformOutputs["aplVersion"] = aplChart.Version

	ctx.Export("apl", AppPlatformOutputs)

	return aplChart, nil
}

func decodeKubeconfig(data map[string]any) (string, error) {
	enc, ok := data["kubeconfig"].(string)
	if !ok {
		return "", errKubeConfig
	}

	k, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}

	return string(k), nil
}
