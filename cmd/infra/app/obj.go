package app

import (
	"fmt"

	"github.com/pulumi/pulumi-linode/sdk/v5/go/linode"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

var (
	LinodeObjKeys    = pulumi.Map{}
	LinodeObjBuckets = pulumi.Map{}
)

func SetupObj(ctx *pulumi.Context) error {
	objKey, err := linode.NewObjectStorageKey(ctx, "pulumiObjKey", &linode.ObjectStorageKeyArgs{
		Label: pulumi.String("ec-infra-key"),
	})
	if err != nil {
		return err
	}

	LinodeObjKeys["accessKey"] = objKey.AccessKey
	LinodeObjKeys["secretKey"] = objKey.SecretKey

	// Get APL bucket prefix from config
	objPrefix := viper.GetString("appPlatform.obj.prefix")
	objRegion := viper.GetString("appPlatform.region")
	objLabels := viper.GetStringSlice("appPlatform.obj.buckets")

	for _, i := range objLabels {
		bucketName := fmt.Sprintf("%s-%s", objPrefix, i)

		bucket, err := linode.NewObjectStorageBucket(ctx, bucketName, &linode.ObjectStorageBucketArgs{
			AccessKey: objKey.AccessKey,
			SecretKey: objKey.SecretKey,
			Region:    pulumi.String(objRegion),
			Label:     pulumi.String(bucketName),
			LifecycleRules: linode.ObjectStorageBucketLifecycleRuleArray{
				&linode.ObjectStorageBucketLifecycleRuleArgs{
					Id:                                 pulumi.String("global-expiration-policy"),
					Enabled:                            pulumi.Bool(true),
					AbortIncompleteMultipartUploadDays: pulumi.Int(5),
					Expiration: &linode.ObjectStorageBucketLifecycleRuleExpirationArgs{
						Days: pulumi.Int(90),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		LinodeObjBuckets[i] = bucket.Label
	}

	ctx.Export("obj", LinodeObjKeys)
	ctx.Export("objBuckets", LinodeObjBuckets)

	return nil
}
