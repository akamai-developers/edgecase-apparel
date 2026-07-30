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

	"github.com/spf13/viper"
)

type StackRef struct {
	Stack *pulumi.StackReference
}

type PulumiConfig struct {
	Context  map[string]string   `mapstructure:"context"`
	Projects []PulumiStackConfig `mapstructure:"projects"`
}

type PulumiStackConfig struct {
	Name    string `mapstructure:"name"`
	Project string `mapstructure:"project"`
	Stack   string `mapstructure:"stack"`
}

type AplConfig struct {
	Chart        string `mapstructure:"chart,omitempty"`
	Domain       string `mapstructure:"domain,omitempty"`
	Name         string `mapstructure:"name,omitempty"`
	PlatformName string `mapstructure:"platformName,omitempty"`
	Repo         string `mapstructure:"repo,omitempty"`
	Region       string `mapstructure:"region,omitempty"`
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

func (s *StackRef) Get(key string) (any, error) {
	info, err := s.Stack.GetOutputDetails(key)
	if err != nil {
		return nil, err
	}

	if info.SecretValue != nil {
		return info.SecretValue, nil
	}

	return info.Value, nil
}

func (c *AplConfig) HelmTemplate(opts ...map[string]any) (string, error) {
	var opt any

	// Merge option maps if more than one was provided
	switch {
	case len(opts) > 1:
		dst := opts[0]
		for _, i := range opts {
			maps.Copy(dst, i)
		}
		opt = dst
	case len(opts) == 1:
		opt = opts[0]
	default:
		opt = nil
	}

	v, err := helmTemplate(c, c.ValuesTpl, opt)
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

func StackRefInit(ctx *pulumi.Context, slug string) (*StackRef, error) {
	var stkRef StackRef

	st, err := pulumi.NewStackReference(ctx, slug, nil)
	if err != nil {
		return nil, err
	}

	stkRef.Stack = st

	return &stkRef, nil
}

func RandInitPass() string {
	return uuid.NewString()
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

func helmTemplate[T HelmConf](tplData T, tpl string, opts any) (string, error) {
	// Make data map for helm values
	// Assign struct values to "main" and optional value maps to "opts"
	data := map[string]any{
		"main": tplData,
	}

	if opts != nil {
		data["opts"] = opts
	}

	projRoot, _ := GetProjRoot("helmTemplate()")
	v := filepath.Join(projRoot, "cmd/templates/helm", tpl)

	funcMap := template.FuncMap{
		"randInitPass": RandInitPass,
	}

	// Execute template against the final data map
	t := template.Must(template.New(tpl).Funcs(funcMap).ParseFiles(v))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, &data); err != nil {
		return "", err
	}

	return buf.String(), nil
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
