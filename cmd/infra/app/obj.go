package app

import (
	"github.com/pulumi/pulumi-linode/sdk/v5/go/linode"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

var LinodeObjKeys, LinodeObjBuckets pulumi.Map

func SetupObj(ctx *pulumi.Context) error {
	objKey, err := linode.NewObjectStorageKey(ctx, "pulumiObjKey", &linode.ObjectStorageKeyArgs{
		Label: pulumi.String("image-access"),
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
		bucketName := objPrefix + i

		bucket, err := linode.NewObjectStorageBucket(ctx, bucketName, &linode.ObjectStorageBucketArgs{
			AccessKey: objKey.AccessKey,
			SecretKey: objKey.SecretKey,
			Region:    pulumi.String(objRegion),
			Label:     pulumi.String(bucketName),
			LifecycleRules: linode.ObjectStorageBucketLifecycleRuleArray{
				&linode.ObjectStorageBucketLifecycleRuleArgs{
					Id:                                 pulumi.String("my-rule"),
					Enabled:                            pulumi.Bool(true),
					AbortIncompleteMultipartUploadDays: pulumi.Int(5),
					Expiration: &linode.ObjectStorageBucketLifecycleRuleExpirationArgs{
						Date: pulumi.String("2021-06-21"),
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
	// ctx.Export("objBuckets", buckets)

	return nil
}
