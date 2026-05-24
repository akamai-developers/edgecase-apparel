# Pulumi IaC

[Pulumi](https://github.com/pulumi/pulumi) is an Infrastructure as Code (IaC) solution that brings both declarative and imperative elements, using SDKs in the familiar programming languages we know and love. Additionally it supports both YAML and HCL configuration/markup languages. This project uses the [Golang SDK](https://github.com/pulumi/pulumi/tree/master/sdk/go) and [Linode provider](https://www.pulumi.com/registry/packages/linode/).

## State Backend

IaC tooling needs a backend for storing state. To remain as vendor-neutral as possible, rather than relying on a fully-managed SaaS, we are self-managing state in an S3 compatible Object Storage bucket. This setup is cloud agnostic―the S3 protocol is standard, and Object Storage is a cloud primitive we can find on any IaaS provider.


Set the `PULUMI_CONFIG_PASSPHRASE` environment variable with a passphrase value that Pulumi can use for encrypting/decrypting state. You'll also need to set the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables. Their values should be stored in a central secrets management vault.

```bash
export PULUMI_CONFIG_PASSPHRASE=${CONFIG_SECRET}
export AWS_ACCESS_KEY_ID=${LINODE_OBJ_ACCESS_KEY}
export AWS_SECRET_ACCESS_KEY=${LINODE_OBJ_SECRET_KEY}
```

The OBJ backend is then configured in `Pulumi.yaml` with the S3 URL. This encoded URL contains the bucket name, endpoint, region, and enables the `disableSSL` and `s3ForcePathStyle` options.  The syntax is `s3://<BUCKET_NAME>?endpoint=<BUCKET_ENDPOINT>?disableSSL=true&s3ForcePathStyle=true&region=<BUCKET_REGION>`.

```yaml
name: edgecase-apparel-infra
description: Akamai EdgeCase Apparel
runtime: go
backend:
  url: 's3://miami-pulumi-backend?endpoint=us-mia-1.linodeobjects.com&disableSSL=true&s3ForcePathStyle=true&region=us-mia'
config:
  linode:objUseTempKeys: true
  linode:objBucketForceDelete: true
  pulumi:tags:
    value:
      pulumi:template: linode-go
```

> [!NOTE]
> The `linode` config tags `objUseTempKeys` and `objBucketForceDelete` are [provider configuration options](https://www.pulumi.com/registry/packages/linode/#configuration-reference) (not Pulumi S3 backend configuration).
>
> - `objUseTempKeys` creates a temporary key pair at apply-time for each bucket operation, and thus reduces the amount of OBJ key pair secretes we have to manage.
> - `objBucketForceDelete` allows Pulumi destroy operations to (allegedly) delete a bucket even when not empty, but purging all its objects beforehand.

