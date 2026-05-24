package internal

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func BuildPulumiStringArray(sa []string) pulumi.StringArray {
	strArray := pulumi.StringArray{}
	for _, i := range sa {
		strArray = append(strArray, pulumi.String(i))
	}

	return strArray
}

func BuildPulumiStringMap(sm map[string]string) pulumi.StringMap {
	strMap := pulumi.StringMap{}
	for k, v := range sm {
		strMap[k] = pulumi.String(v)
	}

	return strMap
}
