package app

import (
	"os"

	cfg "github.com/akamai-developers/edgecase-apparel/cmd/config"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-null/sdk/go/null"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	ptime "github.com/pulumiverse/pulumi-time/sdk/go/time"
	"github.com/spf13/viper"
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
//nolint:funlen, cyclop
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
	kubeconfig := stkRef.GetKubeConfig("primary").Decode().Write()

	k8sProvider, err := kubernetes.NewProvider(ctx, "ec-kubernetes-provider", &kubernetes.ProviderArgs{
		EnableServerSideApply: pulumi.Bool(true),
		Kubeconfig:            pulumi.String(kubeconfig.Config),
	})
	if err != nil {
		return nil, err
	}

	// Get OBJ keys and bucket labels
	objData, err := stkRef.GetObj()
	if err != nil {
		return nil, err
	}

	// Get Keycloak admin password
	adminSecrets := map[string]any{
		"keycloak": map[string]any{
			"adminPass":    os.Getenv("ECA_KEYCLOAK_ADMIN_PASS"),
			"clientSecret": os.Getenv("ECA_KEYCLOAK_CLIENT_SECRET"),
		},
		"apps": map[string]any{
			"loki": os.Getenv("ECA_LOKI_ADMIN_PASS"),
		},
		"teams": map[string]any{
			"core":   os.Getenv("ECA_CORE_TEAM_PASS"),
			"sysdev": os.Getenv("ECA_SYSDEV_TEAM_PASS"),
		},
		"git": map[string]any{
			"ghtoken": os.Getenv("ECA_GH_TOKEN"),
		},
	}

	// Override Helm values template variables
	values, err := apl.HelmTemplate(objData, adminSecrets)
	if err != nil {
		return nil, err
	}

	// var opt = pulumi.ResourceOption{}

	// Deploy APL Helm chart
	opts := []pulumi.ResourceOption{
		pulumi.Provider(k8sProvider),
		pulumi.DeletedWith(k8sProvider),
		pulumi.IgnoreChanges([]string{"version", "values"}),
	}

	aplChart, err := helm.NewRelease(ctx, apl.Name, &helm.ReleaseArgs{
		Chart:           pulumi.String(apl.Chart),
		CreateNamespace: pulumi.Bool(true),
		DisableCRDHooks: pulumi.Bool(false),
		DisableWebhooks: pulumi.Bool(false),
		Lint:            pulumi.Bool(true),
		Name:            pulumi.String(apl.Chart),
		RecreatePods:    pulumi.Bool(true),
		RepositoryOpts: helm.RepositoryOptsArgs{
			Repo: pulumi.String(apl.Repo),
		},
		ReuseValues:   pulumi.Bool(true),
		TakeOwnership: pulumi.Bool(true),
		Timeout:       pulumi.Int(1200),
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewStringAsset(values),
		},
		Version:     pulumi.StringPtr(apl.Version),
		WaitForJobs: pulumi.Bool(true),
	}, opts...)
	if err != nil {
		return nil, err
	}

	// Structure Helm chart outputs
	chartRepo := aplChart.RepositoryOpts.Repo()
	chartStatus := aplChart.Status

	// Chain null resources with a sleep timer, to give time for Helm
	// installation to create ConfigMap and Secrets.
	opts = []pulumi.ResourceOption{pulumi.DependsOn([]pulumi.Resource{aplChart})}

	previous, err := null.NewResource(ctx, "previous", nil, opts...)
	if err != nil {
		return nil, err
	}

	wait30Seconds, err := ptime.NewSleep(ctx, "wait5minutes", &ptime.SleepArgs{
		CreateDuration: pulumi.String("8m"),
	}, pulumi.DependsOn([]pulumi.Resource{previous}))
	if err != nil {
		return nil, err
	}

	next, err := null.NewResource(ctx, "next", nil, pulumi.DependsOn([]pulumi.Resource{
		wait30Seconds,
	}))
	if err != nil {
		return nil, err
	}

	// Fetch Kubernetes ConfigMap and Secrets data
	consoleUrl := []KubeConfigMapSecret[*corev1.ConfigMap]{
		{Keys: []string{"consoleUrl"}, Name: "welcome", Namespace: "apl-operator"},
	}

	secretName := "platform-admin-initial-credentials"
	adminLogin := []KubeConfigMapSecret[*corev1.Secret]{
		{Keys: []string{"username", "password"}, Name: secretName, Namespace: "keycloak"},
	}

	// hook, err := ctx.RegisterResourceHook("installerCheck", configMapCheck, nil)
	// if err != nil {
	// 	return nil, err
	// }

	// hook, err := ctx.RegisterErrorHook(
	// 	"retryInitAdminCreds",
	// 	func(args *pulumi.ErrorHookArgs) (bool, error) {
	// 		latest := ""
	// 		if len(args.Errors) > 0 {
	// 			latest = args.Errors[0]
	// 			msg := fmt.Sprintf("\n\n\nTHE ERROR IS: %v\n\n\n", latest)
	// 			fmt.Println(msg)
	// 			ctx.Log.Debug(msg, nil)
	// 		}

	// 		if !strings.Contains(latest, "does not exist") {
	// 			return false, nil
	// 		}

	// 		time.Sleep(30 * time.Second)
	// 		return true, nil
	// 	},
	// )
	// if err != nil {
	// 	// msg := fmt.Sprintf("\n\n\nTHE ERROR IS: %v\n\n\n", err)
	// 	// fmt.Println(msg)
	// 	// ctx.Log.Debug(msg, nil)
	// 	return nil, err
	// }

	initCreds, err := DeployNewKubeSecrets(ctx, "initAdminCreds", &KubeSecretsArgs{
		ConfigMaps: consoleUrl,
		Secrets:    adminLogin,
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{aplChart, next}))
	if err != nil {
		return nil, err
	}

	// Structure initial admin credential outputs
	//
	//nolint:forcetypeassert
	loginInfo := initCreds.Data.ApplyT(func(m map[string]any) map[string]string {
		// Index ConfigMaps and Secrets
		cm := m["configMaps"].(map[string]any)
		sec := m["secrets"].(map[string]any)

		// Format data map
		output := map[string]string{
			"consoleUrl": cm["consoleUrl"].(string),
			"password":   sec["password"].(string),
			"username":   sec["username"].(string),
		}

		return output
	}).(pulumi.StringMapOutput)

	// Combine chart and admin credential outputs
	stkOutputs := pulumi.Map{
		"chart":        chartRepo,
		"status":       chartStatus,
		"initialAdmin": loginInfo,
	}

	// Export outputs
	ctx.Export("apl", stkOutputs)

	return aplChart, nil
}
