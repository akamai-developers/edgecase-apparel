package app

import (
	"strconv"

	cfg "github.com/akamai-developers/edgecase-apparel/cmd/config"
	utils "github.com/akamai-developers/edgecase-apparel/internal"
	"github.com/pulumi/pulumi-linode/sdk/v5/go/linode"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

var LinodeDNSOutputs = pulumi.Map{}

func SetupDNS(ctx *pulumi.Context) (map[string]any, error) {
	// Get Viper configs
	if err := cfg.InitConfig(); err != nil {
		return nil, err
	}

	var dnsConf cfg.DnsConfig
	if err := viper.UnmarshalKey("dns", &dnsConf); err != nil {
		return nil, err
	}

	// Create domain DNS zone
	domain, err := linode.NewDomain(ctx, dnsConf.Domain, &linode.DomainArgs{
		Domain:   pulumi.String(dnsConf.Domain),
		SoaEmail: pulumi.String(dnsConf.Soa),
		Tags:     utils.BuildPulumiStringArray(dnsConf.Tags),
		TtlSec:   pulumi.Int(dnsConf.TtlSec),
		Type:     pulumi.String(dnsConf.Type),
	})
	if err != nil {
		return nil, err
	}

	// Get domain ID
	id := domain.ID().ApplyT(func(v string) int {
		id, _ := strconv.Atoi(v)
		return id
	}).(pulumi.IntOutput)

	// Create CAA record for letsencrypt wildcard certificate
	caaRecord, err := linode.NewDomainRecord(ctx, "wildcardCAA", &linode.DomainRecordArgs{
		DomainId:   id,
		RecordType: pulumi.String("CAA"),
		Target:     pulumi.String("letsencrypt.org"),
		Name:       pulumi.String(""),
		Tag:        pulumi.String("issuewild"),
		TtlSec:     pulumi.Int(30),
	}, pulumi.DependsOn([]pulumi.Resource{domain}))
	if err != nil {
		return nil, err
	}

	// Export outputs
	LinodeDNSOutputs["domain"] = domain.Domain
	LinodeDNSOutputs["domain_id"] = domain.ID()
	LinodeDNSOutputs["cca_record_id"] = caaRecord.ID()

	ctx.Export("dns", LinodeDNSOutputs)

	// Export map of dns resources
	dnsResources := map[string]any{
		"domain":       domain,
		"domainRecord": caaRecord,
	}

	return dnsResources, nil
}
