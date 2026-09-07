package app

import (
	"fmt"
	"path/filepath"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type KubeSecretsMocks int

// NewResource intercepts resource creations for injecting mock values.
func (KubeSecretsMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	if args.TypeToken == "pkg:index:KubeClusterSecrets$kubernetes:core/v1:ConfigMap" {
		outputs := resource.NewPropertyMapFromMap(map[string]any{
			"consoleUrl": "https://console.edgecase-apparel.app",
		})

		// Copy inputs and add outputs
		state := args.Inputs.Copy()
		state["outputs"] = resource.NewObjectProperty(outputs)

		return fmt.Sprintf("%s::%s", args.TypeToken, args.Name), state, nil
	}

	// Handle all other resources
	return fmt.Sprintf("%s::%s", args.TypeToken, args.Name), args.Inputs, nil
}

// Call intercepts data source lookups (Get / Invoke functions) for injecting mock values.
func (KubeSecretsMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if args.Token == "pkg:index:KubeClusterSecrets$kubernetes:core/v1:ConfigMap" {
		data := pulumi.StringMap{
			"consoleUrl": pulumi.String("[unknown]"),
			"welcome":    pulumi.String("welcome to APL"),
		}
		name := "welcome"
		namespace := "apl-operator"
		id := filepath.Join(namespace, name)

		return resource.NewPropertyMapFromMap(map[string]any{
			id: MockKubeConfigMap(name, namespace, data),
		}), nil
	}

	if args.Token == "pkg:index:KubeClusterSecrets$kubernetes:core/v1:Secret" {
		data := pulumi.StringMap{
			"username": pulumi.String("[unknown]"),
			"password": pulumi.String("[unknown]"),
		}
		name := "platform-admin-initial-credentials"
		namespace := "keycloak"
		id := filepath.Join(namespace, name)

		return resource.NewPropertyMapFromMap(map[string]any{
			id: MockKubeSecret(name, namespace, data),
		}), nil
	}

	return args.Args, nil
}

// MockKubeConfigMap returns a Kubernetes ConfigMap resource for mocks.
func MockKubeConfigMap(name, namespace string, data pulumi.StringMap) *corev1.ConfigMap {
	configMap := new(corev1.ConfigMap)
	configMap.Data = data.ToStringMapOutput()
	configMap.Kind = pulumi.String("ConfigMap").ToStringOutput()
	configMap.Metadata = metav1.ObjectMetaArgs{
		Name:      pulumi.String(name),
		Namespace: pulumi.String(namespace),
	}.ToObjectMetaOutput()

	return configMap
}

// MockKubeSecret returns a Kubernetes Secret resource for mocks.
func MockKubeSecret(name, namespace string, data pulumi.StringMap) *corev1.Secret {
	secret := new(corev1.Secret)
	secret.Data = data.ToStringMapOutput()
	secret.Kind = pulumi.String("Secret").ToStringOutput()
	secret.Metadata = metav1.ObjectMetaArgs{
		Name:      pulumi.String(name),
		Namespace: pulumi.String(namespace),
	}.ToObjectMetaOutput()

	return secret
}
