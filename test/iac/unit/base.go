package unit

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	aplRepo      string = "https://linode.github.io/apl-core"
	ccaRecordId  string = "123456"
	domain       string = "edgecase-apparel.app"
	domainId     string = "251816"
	lkeId        string = "123456"
	lkeLabel     string = "edgecase-apparel-main"
	linodeToken  string = "55772f00-85e0-11f1-85a5-b3f1d391c9f3"
	platformName string = "edgecase-apparel-platform"
	objPrefix    string = "ec-apl-"
)

var kubeconfig = loadKubeconfig()

var objKeys = map[string]any{
	"accessKey": "2ab3b96e8ad111f197ae0ba8880fda5c",
	"secretKey": "372fbf768ad111f180be134ebd0c8ed838da6f9c8ad111f18",
}

var objBuckets = map[string]any{
	"loki":   objPrefix + "loki",
	"cnpg":   objPrefix + "cnpg",
	"harbor": objPrefix + "harbor",
	"thanos": objPrefix + "thanos",
	"tempo":  objPrefix + "tempo",
	"gitea":  objPrefix + "gitea",
}

var aplStatus = map[string]any{
	"chart":     "apl",
	"name":      "apl",
	"namespace": "default",
	"status":    "deployed",
}

var intitAdmin = map[string]any{
	"consoleUrl": "https://console." + domain,
	"password":   randPass(25),
	"username":   "platform-admin@" + domain,
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

func NewTestStack(name string) TestStack {
	ts := new(TestStack)
	n := strings.ToUpper(name)
	p := fmt.Sprintf("ECA_%s_PROJECT", n)
	s := fmt.Sprintf("ECA_%s_STACK", n)

	proj := os.Getenv(p)
	stack := os.Getenv(s)

	ts.Name = n
	ts.Project = proj
	ts.Slug = filepath.Join("organization", proj, stack)
	ts.Stack = stack

	return *ts
}

func randPass(n int) string {
	str := uuid.NewString()
	enc := base64.StdEncoding.EncodeToString([]byte(str))
	pass := []rune(enc)

	return string(pass[:n])
}

func loadKubeconfig() string {
	data, err := os.ReadFile("fixtures/kubeconfig.yaml")
	if err != nil {
		log.Fatal(err)
	}

	enc := base64.StdEncoding.EncodeToString(data)

	return enc
}
