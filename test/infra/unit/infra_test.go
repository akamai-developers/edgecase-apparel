package unit

import (
	"strconv"
	"sync"
	"testing"

	"github.com/akamai-developers/edgecase-apparel/cmd/infra/app"
	"github.com/go-openapi/testify/v2/assert"

	"github.com/pulumi/pulumi-linode/sdk/v5/go/linode"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type infraMocks int

func (infraMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	// Handle StackReference resources
	if args.TypeToken == "pulumi:pulumi:StackReference" {
		outputs := resource.NewPropertyMapFromMap(map[string]interface{}{
			"primary": map[string]any{
				"lke_id":    lke_id,
				"lke_label": lke_label,
			},
			"dns": map[string]any{
				"cca_record_id": cca_record_id,
				"domain":        domain,
				"domain_id":     domain_id,
			},
			// "apl": map[string]any{
			// 	"apl_id":      apl_id,
			// 	"apl_status":  apl_status,
			// 	"apl_version": apl_version,
			// },
		})

		// Copy inputs and add outputs
		state := args.Inputs.Copy()
		state["outputs"] = resource.NewObjectProperty(outputs)
		return args.Name + "_id", state, nil
	}

	if args.TypeToken == "linode:index/domain:Domain" {
		return domain_id, args.Inputs, nil
	}
	// Handle all other resources
	return args.Name + "_id", args.Inputs, nil
}

func (infraMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

// var ts TestStack

func TestInfraStackOutputs(t *testing.T) {
	ts := *new(TestStack)
	ts.Init("infra")
	// proj := os.Getenv("ECA_INFRA_PROJECT")
	// stack := os.Getenv("ECA_INFRA_STACK")
	// fmt.Println(proj)
	// fmt.Println(stack)
	// fmt.Printf("\nproj is %s\n", ts.Project)
	// fmt.Printf("\nstack is %s\n", ts.Stack)

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		// slug := "organization/" + project + "/dev"
		err := app.Deploy(ctx)
		assert.NoError(t, err)

		// Get stack reference
		stk, err := pulumi.NewStackReference(ctx, ts.Slug, nil)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)

		// Test LKE cluster stack references
		go func() {
			defer wg.Done()

			outputs, err := stk.GetOutputDetails("primary")
			assert.NoError(t, err)

			if values, ok := outputs.Value.(map[string]any); ok {
				assert.Equal(t, lke_id, values["lke_id"])
				assert.Equal(t, lke_label, values["lke_label"])
			} else {
				t.Error("error: invalid lke data returned from stack reference")
			}
		}()

		// Test DNS stack references
		go func() {
			defer wg.Done()

			outputs, err := stk.GetOutputDetails("dns")
			assert.NoError(t, err)

			if values, ok := outputs.Value.(map[string]any); ok {
				assert.Equal(t, cca_record_id, values["cca_record_id"])
				assert.Equal(t, domain, values["domain"])
				assert.Equal(t, domain_id, values["domain_id"])
			} else {
				t.Error("error: invalid type returned from stack reference")
			}
		}()

		// // Test APL specific stack references
		// go func() {
		// 	defer wg.Done()

		// 	outputs, err := st.GetOutputDetails("apl")
		// 	assert.NoError(t, err)

		// 	if values, ok := outputs.Value.(map[string]any); ok {
		// 		assert.Equal(t, apl_id, values["apl_id"])
		// 		assert.Equal(t, apl_status, values["apl_status"])
		// 		assert.Equal(t, apl_version, values["apl_version"])
		// 	} else {
		// 		t.Error("error: invalid type returned from stack reference")
		// 	}
		// }()

		wg.Wait()
		return nil
	}, pulumi.WithMocks(ts.Project, ts.Stack, infraMocks(0)))
	assert.NoError(t, err)
}

func TestSetupDns(t *testing.T) {
	ts := *new(TestStack)
	ts.Init("infra")

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		dns, err := app.SetupDNS(ctx)
		assert.NoError(t, err)

		ld := dns["domain"].(*linode.Domain)
		ldr := dns["domainRecord"].(*linode.DomainRecord)

		var wg sync.WaitGroup
		wg.Add(4)

		// Test domain name, SOA email, TTL value, and tag
		pulumi.All(ld.URN(), ld.Domain, ld.SoaEmail, ld.TtlSec, ld.Tags).ApplyT(func(all []any) error {
			urn := all[0].(pulumi.URN)
			_domain := all[1].(string)
			soa := all[2].(*string)
			ttl := all[3].(*int)
			tags := all[4].([]string)

			assert.Equalf(t, domain, _domain, "invalid domain: %v", urn)
			assert.Equalf(t, "devrel@akamai.com", *soa, "invalid soa email address: %v", urn)
			assert.Equalf(t, 30, *ttl, "invalid TTL value: %v", urn)
			assert.SliceContainsTf(t, tags, "ec-apparel", "missing 'ec-apparel' tag: %v", urn)
			wg.Done()

			return nil
		})

		// Test CAA record domain ID
		pulumi.All(ldr.URN(), ldr.DomainId, ld.ID()).ApplyT(func(all []any) error {
			urn := all[0].(pulumi.URN)
			domainId := all[1].(int)
			id, _ := strconv.Atoi(domain_id)

			assert.Equalf(t, id, domainId, "invalid or missing domain ID: %v", urn)
			wg.Done()

			return nil
		})

		// Test CAA record name
		pulumi.All(ldr.URN(), ldr.Name).ApplyT(func(all []any) error {
			urn := all[0].(pulumi.URN)
			name := all[1].(string)

			assert.Equalf(t, "", name, "name value should be empty string: %v", urn)
			wg.Done()

			return nil
		})

		// Test CAA record type, target, and tag
		pulumi.All(ldr.URN(), ldr.RecordType, ldr.Target, ldr.Tag, ldr.TtlSec).ApplyT(func(all []any) error {
			urn := all[0].(pulumi.URN)
			rtype := all[1].(string)
			target := all[2].(string)
			tag := all[3].(*string)
			ttl := all[4].(*int)

			assert.Equalf(t, "CAA", rtype, "invalid record type: %v", urn)
			assert.Equalf(t, "letsencrypt.org", target, "invalid record target: %v", urn)
			assert.Equalf(t, "issuewild", *tag, "invalid record tag: %v", urn)
			assert.Equalf(t, 30, *ttl, "invalid record ttl value: %v", urn)
			wg.Done()

			return nil
		})

		wg.Wait()
		return nil
	}, pulumi.WithMocks(ts.Project, ts.Stack, infraMocks(0)))
	assert.NoError(t, err)
}
