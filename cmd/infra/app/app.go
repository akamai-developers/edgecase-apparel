package app

import (
	cfg "github.com/akamai-developers/edgecase-apparel/cmd/config"
	utils "github.com/akamai-developers/edgecase-apparel/internal"

	"github.com/pulumi/pulumi-linode/sdk/v5/go/linode"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var stackOutputMap = pulumi.Map{}

func Deploy(ctx *pulumi.Context) error {
	// Get Viper configs
	if err := cfg.InitConfig(); err != nil {
		return err
	}

	var (
		lke cfg.LkeConfig
		apl cfg.NodePoolConfig
	)

	if err := lke.Get("primary"); err != nil {
		return err
	}

	if err := apl.Get("apl"); err != nil {
		return err
	}

	nodeLabels := apl.MapLabels()

	// Setup DNS zone
	domainResources, err := SetupDNS(ctx)
	if err != nil {
		return err
	}

	domain := domainResources["domain"].(*linode.Domain)

	// Create LKE cluster
	lkeCluster, err := linode.NewLkeCluster(ctx, lke.Label, &linode.LkeClusterArgs{
		AplEnabled: pulumi.Bool(false),
		ControlPlane: &linode.LkeClusterControlPlaneArgs{
			AuditLogsEnabled: pulumi.Bool(lke.ControlPlane.AuditLogs),
			HighAvailability: pulumi.Bool(lke.ControlPlane.HighAvailability),
		},
		K8sVersion: pulumi.String(lke.K8sVersion),
		Label:      pulumi.String(lke.Label),
		Pools: linode.LkeClusterPoolArray{
			&linode.LkeClusterPoolArgs{
				Autoscaler: &linode.LkeClusterPoolAutoscalerArgs{
					Max: pulumi.Int(apl.Autoscaler.Max),
					Min: pulumi.Int(apl.Autoscaler.Min),
				},
				Count:      pulumi.Int(apl.Count),
				K8sVersion: pulumi.String(apl.K8sVersion),
				Label:      pulumi.String(apl.Label),
				Labels:     utils.BuildPulumiStringMap(nodeLabels),
				Tags:       utils.BuildPulumiStringArray(apl.Tags),
				Type:       pulumi.String(apl.Type),
			},
		},
		Region: pulumi.String(lke.Region),
		Tags:   utils.BuildPulumiStringArray(lke.Tags),
	}, pulumi.DependsOn([]pulumi.Resource{domain}))
	if err != nil {
		return err
	}

	stackOutputMap["lke_id"] = lkeCluster.ID()
	stackOutputMap["lke_label"] = lkeCluster.Label

	ctx.Export("primary", stackOutputMap)

	return nil
}
