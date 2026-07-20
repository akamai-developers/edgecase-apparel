package app_test

import (
	"testing"

	"github.com/akamai-developers/edgecase-apparel/cmd/infra/app"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type mocks int

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	// Handle StackReference resources
	if args.TypeToken == "pulumi:pulumi:StackReference" {
		outputs := resource.NewPropertyMapFromMap(map[string]interface{}{
			"primary": map[string]any{
				"lke_id":    "123456",
				"lke_label": "edgecase-apparel-main",
			},
		})

		// Copy inputs and add outputs
		state := args.Inputs.Copy()
		state["outputs"] = resource.NewObjectProperty(outputs)
		return args.Name + "_id", state, nil
	}
	// Handle all other resources
	return args.Name + "_id", args.Inputs, nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func TestDeploy(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		err := app.Deploy(ctx)
		assert.NoError(t, err)

		// Get stack reference
		st, err := pulumi.NewStackReference(ctx, "organization/edgecase-apparel-infra/dev", nil)
		assert.NoError(t, err)

		outputs, err := st.GetOutputDetails("primary")
		assert.NoError(t, err)

		if values, ok := outputs.Value.(map[string]any); ok {
			assert.Equal(t, "123456", values["lke_id"])
			assert.Equal(t, "edgecase-apparel-main", values["lke_label"])
		} else {
			t.Error("error: invalid type returned from stack reference")
		}

		return nil
	}, pulumi.WithMocks("edgecase-apparel-infra", "dev", mocks(0)))
	assert.NoError(t, err)
}
