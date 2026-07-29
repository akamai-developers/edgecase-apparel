package config

import (
	"bytes"
	"fmt"
	"html/template"
	"maps"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	// "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type StackRef struct {
	Stack *pulumi.StackReference
}

// type StackRefVal[T any] struct {
// 	Value T
// }

// type MyStk[T pulumi.StackReference] struct{
// 	Stack T
// }

// type Nah struct {
// 	Details *pulumi.StackReferenceOutputDetails
// 	Key     string
// 	Slug    string
// 	Stack   *pulumi.StackReference
// }

// type StackRefData interface {
// 	*pulumi.StackReference
// 	// GetMap(string) (map[string]string, error)
// 	// GetDeets(*pulumi.Context, string) (map[string]string, error)
// }

// type StackReference[T StackRef] interface {
// 	Get(string) (T, error)
// }

type PulumiConfig struct {
	Context  map[string]string   `mapstructure:"context"`
	Projects []PulumiStackConfig `mapstructure:"projects"`
}

type PulumiStackConfig struct {
	Name    string `mapstructure:"name"`
	Project string `mapstructure:"project"`
	Stack   string `mapstructure:"stack"`
}

type ObjConfig struct {
	Buckets map[string]string `mapstructure:"buckets,omitempty"`
	Keys    map[string]string `mapstructure:"keys,omitempty"`
	Prefix  string            `mapstructure:"prefix,omitempty"`
	// Region string            `mapstructure:"region,omitempty"`
}

// type ObjKeys struct {
// 	AccessKey string `mapstructure:"accessKey,omitempty"`
// 	SecretKey string `mapstructure:"secretKey,omitempty"`
// }

type AplConfig struct {
	Chart  string `mapstructure:"chart,omitempty"`
	Domain string `mapstructure:"domain,omitempty"`
	Name   string `mapstructure:"name,omitempty"`
	// Obj    struct {
	// 	Prefix  string   `mapstructure:"prefix,omitempty"`
	// 	Buckets []string `mapstructure:"buckets,omitempty"`
	// } `mapstructure:"obj,omitempty"`
	// Obj    ObjConfig `mapstructure:"obj,omitempty"`
	// ObjPrefix    string    `mapstructure:"objPrefix,omitempty"`
	PlatformName string `mapstructure:"platformName,omitempty"`
	Repo         string `mapstructure:"repo,omitempty"`
	Region       string `mapstructure:"repo,omitempty"`
	Token        string `mapstructure:"token,omitempty"`
	ValuesTpl    string `mapstructure:"valuesTpl,omitempty"`
	Version      string `mapstructure:"version,omitempty"`
}

type DnsConfig struct {
	Domain string   `mapstructure:"domain,omitempty"`
	Soa    string   `mapstructure:"soa,omitempty"`
	Type   string   `mapstructure:"type,omitempty"`
	Tags   []string `mapstructure:"tags,omitempty"`
	TtlSec int      `mapstructure:"ttlSec,omitempty"`
}

type LkeConfig struct {
	ControlPlane struct {
		AuditLogs        bool `mapstructure:"auditLogs,omitempty"`
		HighAvailability bool `mapstructure:"highAvailability,omitempty"`
	} `mapstructure:"controlPlane,omitempty"`
	K8sVersion string   `mapstructure:"k8sVersion,omitempty"`
	Label      string   `mapstructure:"label,omitempty"`
	Region     string   `mapstructure:"region,omitempty"`
	Tags       []string `mapstructure:"tags,omitempty"`
}

type NodePoolConfig struct {
	Autoscaler struct {
		Min int `mapstructure:"min,omitempty"`
		Max int `mapstructure:"max,omitempty"`
	} `mapstructure:"autoscaler,omitempty"`
	Count      int                `mapstructure:"count,omitempty"`
	K8sVersion string             `mapstructure:"k8sVersion,omitempty"`
	Label      string             `mapstructure:"label,omitempty"`
	Labels     []NodeLabelsConfig `mapstructure:"labels,omitempty"`
	Tags       []string           `mapstructure:"tags,omitempty"`
	Type       string             `mapstructure:"type,omitempty"`
}

type NodeLabelsConfig struct {
	Key   string `mapstructure:"key,omitempty"`
	Value string `mapstructure:"value,omitempty"`
}

type Conf interface {
	*NodePoolConfig | *LkeConfig | *DnsConfig
}

type HelmConf interface {
	*AplConfig
}

func (s *StackRef) GetMap(key string) (map[string]string, error) {
	info, err := s.Stack.GetOutputDetails(key)
	if err != nil {
		return nil, err
	}

	valueMap, err := getMapValue(key, info)
	if err != nil {
		return nil, err
	}

	return valueMap, nil
}

func (c *AplConfig) HelmTemplate(opts ...map[string]string) (string, error) {
	v, err := helmTemplate(c, c.ValuesTpl, opts...)
	return v, err
}

func (c *LkeConfig) Get(key string) error {
	err := unmarshalSubkey(key, c)
	return err
}

func (c *NodePoolConfig) Get(key string) error {
	err := unmarshalSubkey(key, c)
	return err
}

func (c *NodePoolConfig) MapLabels() map[string]string {
	if len(c.Labels) > 0 {
		m := make(map[string]string)
		for _, i := range c.Labels {
			m[i.Key] = i.Value
		}

		return m
	}

	return nil
}

// type helmValueOpts interface {
// 	map[string]string | map[string]any
// }

func StackRefInit(ctx *pulumi.Context, slug string) (*StackRef, error) {
	var stkRef StackRef

	st, err := pulumi.NewStackReference(ctx, slug, nil)
	if err != nil {
		return nil, err
	}

	stkRef.Stack = st

	return &stkRef, nil
}

func getMapValue(key string, info *pulumi.StackReferenceOutputDetails) (map[string]string, error) {
	var i any
	m := make(map[string]string)

	if info.Value != nil {
		i = info.Value
	} else {
		i = info.SecretValue
	}

	switch val := i.(type) {
	case string:
		m[key] = val
		return m, nil
	case map[string]any:
		for k, v := range val {
			if _, ok := v.(string); !ok {
				err := fmt.Errorf("\n[ error ] map item for %s is not string type\n", key)
				return nil, err
			}
			m[k] = v.(string)
		}
		return m, nil
	default:
		err := fmt.Errorf("\n[ error ] %s is not string type: %v\n", key, val)
		return nil, err
	}
}

func helmTemplate[T HelmConf](tplData T, tpl string, opts ...map[string]string) (string, error) {
	var data map[string]string

	// Marshal struct values to yaml []byte string
	yamlData, err := yaml.Marshal(tplData)
	if err != nil {
		fmt.Println("YESYYYYYYYY")
		return "", err
	}

	// Unmarshal yaml string to map
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return "", err
	}

	// Merge with optional maps
	if len(opts) > 0 {
		for _, i := range opts {
			maps.Copy(data, i)
		}
	}

	projRoot, _ := GetProjRoot("helmTemplate()")
	v := filepath.Join(projRoot, "cmd/templates/helm", tpl)

	funcMap := template.FuncMap{
		"randInitPass": RandInitPass,
		"mapValue":     tplMapValue,
	}

	t := template.Must(template.New(tpl).Funcs(funcMap).ParseFiles(v))
	buf := &bytes.Buffer{}

	if err := t.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RandInitPass() string {
	return uuid.NewString()
}

func tplMapValue(m map[string]string, k string) string {
	s := m[k]
	return s
}

func unmarshalSubkey[T Conf](key string, cfg T) error {
	var section string

	switch any(cfg).(type) {
	case *LkeConfig:
		section = "lkeclusters"
	case *NodePoolConfig:
		section = "nodepools"
	}

	result := viper.GetStringMap(section)
	_, ok := result[key]
	if !ok {
		return fmt.Errorf("viper.GetStringMap error: no data (empty map)")
	}

	subKey := fmt.Sprintf("%s.%s", section, key)
	subViper := viper.Sub(subKey)

	if err := subViper.Unmarshal(&cfg); err != nil {
		return err
	}

	return nil
}

func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// viper.AddConfigPath(".")

	// Get absolute path to project root
	projRoot, err := GetProjRoot("InitConfig()")
	if err != nil {
		return err
	}

	// Add project root to config path
	viper.AddConfigPath(projRoot)

	// Look for environment variables prefixed with "ECA"
	viper.SetEnvPrefix("eca")

	// Load any environment variable with the above prefix
	viper.AutomaticEnv()

	// Rename config key/value paths to environment variable delimiter (http.port -> HTTP_PORT)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Look for "ECA_LINODE_TOKEN" environment variable
	if err := viper.BindEnv("linode.token"); err != nil {
		return err
	}

	// Read from the config file
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func GetProjRoot(funcName string) (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("%s func error: runtime caller failed to get source file path", funcName)
	}

	dir := filepath.Dir(filename)
	projRoot := filepath.Join(dir, "..", "..")

	return projRoot, nil
}
