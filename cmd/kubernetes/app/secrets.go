package app

import (
	"encoding/base64"
	"fmt"
	"log"
	"reflect"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

var (
	ConfigMapOutput = pulumi.Map{}
	SecretsOutput   = pulumi.Map{}
)

// KubeSecrets holds a Pulumi component resource state and outputs.
type KubeSecrets struct {
	pulumi.ResourceState

	Data pulumi.MapOutput `pulumi:"kubeClusterSecrets"`
}

// KubeSecretsArgs takes inputs for a Pulumi component resource.
type KubeSecretsArgs struct {
	ConfigMaps []KubeConfigMapSecret[*corev1.ConfigMap] `pulumi:"kubeConfigMaps"`
	Secrets    []KubeConfigMapSecret[*corev1.Secret]    `pulumi:"kubeSecrets"`
}

// KubeConfigMapSecret is a generic type for Kubernetes ConfigMap and Secret
// resources.
type KubeConfigMapSecret[T KubeData] struct {
	Keys      []string
	Kind      T
	Name      string
	Namespace string
}

// GetConfigMapSecret holds values for ConfigMap and Secret <resource>.Get()
// receiver methods.
type GetConfigMapSecret struct {
	Name      string
	Namespace string
	Parent    pulumi.Resource
	Resource  string
}

// KubeData is a generic interface that is constrained to ConfigMap and Secret
// Pulumi Kubernetes types.
type KubeData interface {
	*corev1.ConfigMap | *corev1.Secret
}

// KubeKinds is type constrained interface matching types with ParseData
// receiver methods.
type KubeKinds[T KubeData] interface {
	ParseData()
}

// ParseData extracts map values from the data field of Kubernetes ConfigMap and
// Secret resources.
func (k *KubeConfigMapSecret[T]) ParseData() {
	parseData(k)
}

// KubeConfigMap is a resource Get method for fetching ConfigMap resources.
func (v *GetConfigMapSecret) KubeConfigMap(ctx *pulumi.Context) (*corev1.ConfigMap, error) {
	return getConfigMap(ctx, *v)
}

// KubeSecret is a resource Get method for fetching Secret resources.
func (v *GetConfigMapSecret) KubeSecret(ctx *pulumi.Context) (*corev1.Secret, error) {
	return getSecret(ctx, *v)
}

// DeployNewKubeSecrets is a wrapper for NewKubeSecrets component resource. The
// component is wrapped for the purpose package-level error handling, and to
// workaround failing previews that result from <resource>.Get() method calls to
// resources not created yet. See https://github.com/pulumi/pulumi/issues/3364.
// KubeSecret() and KubeConfigMap() are two such Get() methods. Their requested
// objects are mocked until the real objects are created, allowing previews to
// finish without errors. A successful deployment sets the kubeSecrets stack
// config value, at which point all future previews query the real objects.
func DeployNewKubeSecrets(ctx *pulumi.Context, name string, args *KubeSecretsArgs,
	opts ...pulumi.ResourceOption,
) (*KubeSecrets, error) {
	resource := &KubeSecrets{}

	// Get kubeSecrets stak config value
	cfg := config.New(ctx, "")
	hasKubeSecrets := cfg.Get("kubeSecrets")

	// If preview and kubeSecrets not "true" then mock ConfigMap and Secrets
	if ctx.DryRun() && hasKubeSecrets != "true" {
		err := pulumi.RunErr(func(ctx *pulumi.Context) error {
			kubeSecret, err := NewKubeSecrets(ctx, name, args, opts...)
			resource = kubeSecret

			return err
		}, pulumi.WithMocks("edgecase-apparel-kubernetes", "dev", KubeSecretsMocks(0)))
		if err != nil {
			return nil, err
		}

		return resource, err
	}

	// Otherwise run against actual Kubernetes resources
	return NewKubeSecrets(ctx, name, args, opts...)
}

// NewKubeSecrets is a component resource for indexing values from the data key in
// Kubernetes ConfigMaps and Secrets. As component resource, it has its own state
// that is tracked by Pulumi, and can leverage Pulumi resource options such as
// provider, dependsOn, and hooks, in order to control its execution flow.
func NewKubeSecrets(ctx *pulumi.Context, name string, args *KubeSecretsArgs,
	opts ...pulumi.ResourceOption,
) (*KubeSecrets, error) {
	var resource KubeSecrets

	err := ctx.RegisterComponentResource("pkg:index:KubeClusterSecrets", name, &resource, opts...)
	if err != nil {
		return nil, err
	}

	// Parse data from configmaps
	for _, cfgmap := range args.ConfigMaps {
		data := GetConfigMapSecret{
			Name:      cfgmap.Name,
			Namespace: cfgmap.Namespace,
			Parent:    &resource,
			Resource:  "cm-" + cfgmap.Name,
		}

		configMap, err := data.KubeConfigMap(ctx)
		if err != nil {
			return nil, err
		}

		cfgmap.Kind = configMap
		cfgmap.ParseData()
	}

	// Parse data from secrets
	for _, sec := range args.Secrets {
		data := GetConfigMapSecret{
			Name:      sec.Name,
			Namespace: sec.Namespace,
			Parent:    &resource,
			Resource:  "sec-" + sec.Name,
		}

		secret, err := data.KubeSecret(ctx)
		if err != nil {
			return nil, err
		}

		sec.Kind = secret
		sec.ParseData()
	}

	outputMap := pulumi.Map{
		"configMaps": ConfigMapOutput,
		"secrets":    SecretsOutput,
	}

	resource.Data = outputMap.ToMapOutput()

	err = ctx.RegisterResourceOutputs(&resource, pulumi.Map{})
	if err != nil {
		return nil, err
	}

	return &resource, nil
}

// getConfigMap utilizes the Kubernetes core/v1 Get() method for fetching configmaps
func getConfigMap(ctx *pulumi.Context, obj GetConfigMapSecret) (*corev1.ConfigMap, error) {
	id := fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)

	var state corev1.ConfigMapState

	configMap, err := corev1.GetConfigMap(ctx, obj.Resource, pulumi.ID(id), &state, pulumi.Parent(obj.Parent))
	if err != nil {
		return nil, err
	}

	return configMap, nil
}

// getSecret utilizes the Kubernetes core/v1 Get() method for fetching secrets
func getSecret(ctx *pulumi.Context, obj GetConfigMapSecret) (*corev1.Secret, error) {
	id := fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)

	var secretState corev1.SecretState

	secret, err := corev1.GetSecret(ctx, obj.Resource, pulumi.ID(id), &secretState, pulumi.Parent(obj.Parent))
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// parseData generic function for types that satisfy the KubeData interface. It
// uses reflection to determine the true underlying type, which is constrained
// to *corev1.ConfigMap and *corev1.Secrets. Values parsed from data maps of
// ConfigMap or Secrets objects, can then be exported as stack outputs or
// consumed by other Pulumi resources.
func parseData[T KubeData](kubeData *KubeConfigMapSecret[T]) {
	// Get struct metadata
	typ := reflect.TypeOf(kubeData.Kind)

	// Get struct values
	val := reflect.ValueOf(kubeData.Kind)

	// Dereference pointers
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	// Ensure the data field exists
	_, ok := typ.FieldByName("Data")
	if !ok {
		log.Fatalf("error ParseData(): configMap or secret data attributes not found")
	}

	field := val.FieldByName("Data").Interface()

	data, ok := field.(pulumi.StringMapOutput)
	if !ok {
		log.Fatal("error ParseData(): interface{} is not pulumi.StringMapOutput")
	}

	for _, key := range kubeData.Keys {
		switch typ.Name() {
		case "ConfigMap":
			ConfigMapOutput[key] = ToMappable(data, key)
		case "Secret":
			SecretsOutput[key] = ToMappable(data, key, true)
		}
	}
}

// ToMappable indexes keys from a Pulumi.StringMapOutput interface that is
// returned by the FieldByName method of reflect.Value. Setting the optional
// unmask parameter to 'true' will decode a base64 encoded Kubernets secret and
// prevent Pulumi from masking it in stdout.
func ToMappable(data pulumi.StringMapOutput, key string, unmask ...bool) pulumi.Output {
	item := data.MapIndex(pulumi.String(key))

	if len(unmask) > 0 && unmask[0] {
		rawOutput := pulumi.Unsecret(pulumi.ToOutput(item))
		//nolint:forcetypeassert
		rawStr := rawOutput.ApplyT(func(s string) string {
			decoded, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				log.Fatalf("error toMappable(): %v", err)
			}

			return string(decoded)
		}).(pulumi.StringOutput)

		return pulumi.ToOutput(rawStr)
	}

	return pulumi.ToOutput(item)
}
