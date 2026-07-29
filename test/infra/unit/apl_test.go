package unit

import (
	"fmt"
	"sync"
	"testing"

	"github.com/akamai-developers/edgecase-apparel/cmd/kubernetes/app"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"go.yaml.in/yaml/v3"
)

type aplMocks int

func (aplMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	// Handle StackReference resources
	if args.TypeToken == "pulumi:pulumi:StackReference" {
		outputs := resource.NewPropertyMapFromMap(map[string]interface{}{
			"apl": map[string]any{
				"apl_id":     apl_id,
				"apl_status": apl_status,
			},
			"obj": map[string]any{
				"accessKey": access_key,
				"secretKey": secret_key,
			},
			"objBuckets": objBuckets,
		})

		// Copy inputs and add outputs
		state := args.Inputs.Copy()
		state["outputs"] = resource.NewObjectProperty(outputs)
		return args.Name + "_id", state, nil
	}
	// Handle all other resources
	return args.Name + "_id", args.Inputs, nil
}

func (aplMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func TestDeployApl(t *testing.T) {
	ts := *new(TestStack)
	ts.Init("apl")

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		apl, err := app.DeployApl(ctx)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(3)

		// Test chart and release names, and that linting is enabled
		pulumi.All(apl.URN(), apl.Chart, apl.Lint, apl.Name).ApplyT(func(all []any) error {
			urn := all[0].(pulumi.URN)
			chart := all[1].(string)
			lint := all[2].(*bool)
			name := all[3].(*string)

			assert.Equalf(t, "apl", chart, "invalid chart name: %v", urn)
			assert.Equalf(t, true, *lint, "linting should be 'true': %v", urn)
			assert.Equalf(t, "apl", *name, "invalid release name: %v", urn)
			wg.Done()

			return nil
		})

		// Test version and remote chart repo
		pulumi.All(apl.URN(), apl.RepositoryOpts.Repo()).ApplyT(func(all []any) error {
			urn := all[0].(pulumi.URN)
			repo := all[1].(*string)

			assert.Equalf(t, "https://linode.github.io/apl-core", *repo, "invalid remote chart repo: %v", urn)
			wg.Done()

			return nil
		})

		// Test helm values
		pulumi.All(apl.URN(), apl.ValueYamlFiles.Index(pulumi.Int(0))).ApplyT(func(all []any) error {
			urn := all[0].(pulumi.URN)
			values := all[1].(pulumi.Asset).Text()
			fmt.Println(values)

			var data map[string]any

			err := yaml.Unmarshal([]byte(values), &data)
			assert.NoError(t, err)

			// Test top-level map keys
			expectedKeys := []string{
				"apps",
				"cluster",
				"dns",
				"otomi",
				"teamConfig",
			}

			gotKeys := make([]string, 0)
			for k := range data {
				gotKeys = append(gotKeys, k)
			}

			for _, key := range expectedKeys {
				assert.SliceContainsTf(t, gotKeys, key, "missing required key in helm values: %v", urn)
			}

			// Test correctness of values populated by template variables
			for _, i := range gotKeys {
				submap := data[i].(map[string]any)
				switch i {
				case "app":
					// Test cert-manager domain email
					key := "cert-manager"
					wants := "admin@" + domain
					got := submap[key].(map[string]any)

					assert.Containsf(t, got, wants, "invalid helm value for '%s.%s.%s': %v", i, key, wants, urn)
				case "cluster":
					// Test platform name
					key := "name"
					got := submap[key]
					wants := platformName
					assert.Containsf(t, got, wants, "missing or invalid value for %s.%s: %v", i, key, urn)

					// Test domain suffix
					key = "domainSuffix"
					got = submap[key]
					wants = domain
					assert.Containsf(t, got, wants, "missing or invalid value for %s.%s: %v", i, key, urn)
				case "dns":
					// Test ExtenalDNS domain filter
					key := "domainFilters"
					got := submap[key]
					assert.Containsf(t, got, domain, "missing or invalid values for %s.%s: %v", i, key, urn)

					// Test Linode API token value
					key = "provider"
					nested := submap[key].(map[string]any)
					yamlData, err := yaml.Marshal(nested)
					assert.NoError(t, err)

					tokenData := struct {
						Linode struct {
							ApiToken string `yaml:"apiToken,omitempty"`
						} `yaml:"linode,omitempty"`
					}{}

					err = yaml.Unmarshal(yamlData, &tokenData)
					assert.NoError(t, err)

					got = tokenData.Linode.ApiToken
					wants := linode_token
					assert.Containsf(t, got, wants, "missing value for %s.%s.linode.apiToken: %v", i, key, urn)
				case "otomi":
					// Test that adminPassword contains a value
					key := "adminPassword"
					got := submap[key]
					assert.NotZerof(t, got, "missing %s.%s value: %v", i, key, urn)
				}
			}

			wg.Done()
			return nil
		})

		wg.Wait()
		return nil
	}, pulumi.WithMocks(ts.Project, ts.Stack, aplMocks(0)))
	assert.NoError(t, err)
}
