package unit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	apl_id     string = "452341"
	apl_status string = "deployed"

	cca_record_id string = "123456"
	domain        string = "edgecase-apparel.app"
	domain_id     string = "251816"

	lke_id    string = "123456"
	lke_label string = "edgecase-apparel-main"

	linode_token string = "55772f00-85e0-11f1-85a5-b3f1d391c9f3"
	platformName string = "edgecase-apparel-platform"

	objPrefix  string = "ec-apl-"
	access_key string = "2ab3b96e8ad111f197ae0ba8880fda5c"
	secret_key string = "372fbf768ad111f180be134ebd0c8ed838da6f9c8ad111f18"
)

var objBuckets = map[string]any{
	"loki":   objPrefix + "loki",
	"cnpg":   objPrefix + "cnpg",
	"harbor": objPrefix + "harbor",
	"thanos": objPrefix + "thanos",
	"tempo":  objPrefix + "tempo",
	"gitea":  objPrefix + "gitea",
}

type TestStack struct {
	Name    string
	Project string
	Slug    string
	Stack   string
}

func (t *TestStack) Init(name string) {
	n := strings.ToUpper(name)
	p := fmt.Sprintf("ECA_%s_PROJECT", n)
	s := fmt.Sprintf("ECA_%s_STACK", n)

	proj := os.Getenv(p)
	stack := os.Getenv(s)

	t.Name = n
	t.Project = proj
	t.Slug = filepath.Join("organization", proj, stack)
	t.Stack = stack
}
