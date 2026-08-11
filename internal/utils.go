// Package internal provides shared utility functions for common tasks across
// importable public modules.
package internal

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// BuildPulumiStringArray rebuilds a []string value as an array of
// pulumi.String() typed values. This is useful for inline conversion, for
// Pulumi resource inputs.
func BuildPulumiStringArray(sa []string) pulumi.StringArray {
	strArray := pulumi.StringArray{}
	for _, i := range sa {
		strArray = append(strArray, pulumi.String(i))
	}

	return strArray
}

// BuildPulumiStringMap rebuilds a map[string]string value as a
// map[string]pulumi.String() type. This is useful for inline conversion, for
// Pulumi resource inputs.
func BuildPulumiStringMap(sm map[string]string) pulumi.StringMap {
	strMap := pulumi.StringMap{}
	for k, v := range sm {
		strMap[k] = pulumi.String(v)
	}

	return strMap
}
