package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spf13/viper"
)

// StackRef holds the exported outputs from a Pulumi stack reference.
type StackRef struct {
	Stack *pulumi.StackReference
}

// PulumiConfigs is an array of Pulumi stacks defined in config.yaml.
type PulumiConfigs struct {
	Projects []struct {
		Name    string `mapstructure:"name"`
		Project string `mapstructure:"project"`
		Stack   string `mapstructure:"stackCtx"`
	} `mapstructure:"pulumi"`
}

// AplConfig receives config file values which are required for APL helm chart
// parameters and values.yaml file overrides.
type AplConfig struct {
	Chart  string `mapstructure:"chart,omitempty"`
	Domain string `mapstructure:"domain,omitempty"`
	// Kubeconfig   string `mapstructure:"kubeconfig,omitempty"`
	Name         string `mapstructure:"name,omitempty"`
	PlatformName string `mapstructure:"platformName,omitempty"`
	Repo         string `mapstructure:"repo,omitempty"`
	Region       string `mapstructure:"region,omitempty"`
	Token        string `mapstructure:"token,omitempty"`
	ValuesTpl    string `mapstructure:"valuesTpl,omitempty"`
	Version      string `mapstructure:"version,omitempty"`
}

// DnsConfig receives config file values that define the DNS zone to be created.
type DnsConfig struct {
	Domain string   `mapstructure:"domain,omitempty"`
	Soa    string   `mapstructure:"soa,omitempty"`
	Type   string   `mapstructure:"type,omitempty"`
	Tags   []string `mapstructure:"tags,omitempty"`
	TtlSec int      `mapstructure:"ttlSec,omitempty"`
}

// LkeConfig receives config file values that define all LKE cluster parameters
// other than node pools.
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

// NodePoolConfig receives config file values that define a node pool to add or
// provision with the LKE cluster.
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

// NodeLabelsConfig represents a Kubernetes key-value label for nodes within a
// particular node pool.
type NodeLabelsConfig struct {
	Key   string `mapstructure:"key,omitempty"`
	Value string `mapstructure:"value,omitempty"`
}

// Conf is generic interface for DnsConfig, LkeConfig, and NodePoolConfig types.
// This is to avoid coding repetitive functions that perform the same task, by
// enabling use of single generic functions that satisfy all of these types.
type Conf interface {
	*NodePoolConfig | *LkeConfig | *DnsConfig
}

// HelmConf is a generic interface for types that represent values for Helm
// chart parameters.
type HelmConf interface {
	*AplConfig
}

type Kubeconfig struct {
	Config string
	Data   map[string]any
	Name   string
}

// Get retrieves an exported value from a Pulumi stack reference.
//
//nolint:ireturn
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

// HelmTemplate is a receiver method that returns a string representation of
// the chart values.yaml file, populated from AplConfig fields. When provided,
// values from optional map[string]any types are appended.
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

// Get unmarshals LKE cluster values from YAML config.
func (c *LkeConfig) Get(key string) error {
	err := unmarshalSubkey(key, c)

	return err
}

// Get unmarshals LKE cluster values from YAML config.
func (c *NodePoolConfig) Get(key string) error {
	err := unmarshalSubkey(key, c)

	return err
}

// MapLabels iterates through an array of NodeLabelsConfig types and returns a
// map[string]string value.
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

// GetObj fetches the value of OBJ keys and bucket labels exported from a
// Pulumi stack reference, performs type checking validation, and then returns
// a map[string]any value.
func (s *StackRef) GetObj() (map[string]any, error) {
	objKeys, err := s.Get("obj")
	if err != nil {
		return nil, err
	}

	objBuckets, err := s.Get("objBuckets")
	if err != nil {
		return nil, err
	}

	opts := map[string]any{
		"objKeys":    objKeys,
		"objBuckets": objBuckets,
	}

	return opts, nil
}

// GetKubeConfig fetches the base64 string value of the LKE cluster kubeconfig
// from an exported Pulumi stack reference, and returns a map[string]any value.
func (s *StackRef) GetKubeConfig(key string) *Kubeconfig {
	return getKubeConfig(s, key)
}

func (k *Kubeconfig) Decode() *Kubeconfig {
	return decodeKubeconfig(k)
}

func (k *Kubeconfig) Write() *Kubeconfig {
	return writeKubeconfig(k)
}

// StackRefInit initializes a Pulumi stack reference for a given stack name.
func StackRefInit(ctx *pulumi.Context, stack string) (*StackRef, error) {
	var (
		stkRef      StackRef
		pulumiConfs PulumiConfigs
	)

	if err := viper.UnmarshalKey("pulumi", &pulumiConfs.Projects); err != nil {
		return nil, err
	}

	var slug string

	for _, i := range pulumiConfs.Projects {
		if i.Name == stack {
			slug = filepath.Join("organization", i.Project, i.Stack)
		}
	}

	if slug == "" {
		err := fmt.Errorf("[ error ] invalid stack name argument: %s", stack)

		return nil, err
	}

	st, err := pulumi.NewStackReference(ctx, slug, nil)
	if err != nil {
		return nil, err
	}

	stkRef.Stack = st

	return &stkRef, nil
}

// RandInitPass is a function for generating random 40 character passwords.
func RandInitPass() string {
	str := uuid.NewString()
	enc := base64.StdEncoding.EncodeToString([]byte(str))
	pass := []rune(enc)

	return string(pass[:40])
}

// InitConfig is a function imported from the Viper library that reads in values
// from a specified config file or prefixed environment variables.
// See: https://github.com/spf13/viper#reading-config-files
func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// viper.AddConfigPath("../../")

	// Get absolute path to project root
	projRoot, err := GetProjRoot("InitConfig()")
	if err != nil {
		return err
	}

	// Add project root to config path
	viper.AddConfigPath(projRoot)
	// cwd, err := os.Getwd()
	// if err != nil {
	// 	return err
	// }
	// path := filepath.Join("../", cwd)
	// viper.AddConfigPath(path)

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

// GetProjRoot returns the relative path of the project root.
func GetProjRoot(funcName string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		// Move up one directory level unti reaching root.
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errGetProjRoot
		}

		dir = parent
	}
}

func getKubeConfig(stkref *StackRef, key string) *Kubeconfig {
	if key != "primary" && key != "backup" {
		log.Fatal(errInvalidKey)
	}

	val, err := stkref.Get(key)
	if err != nil {
		log.Fatal(errStkRefNoValue)
	}

	data, ok := val.(map[string]any)
	if !ok {
		log.Fatal(errStkRefType)
	}

	label := fmt.Sprintf("lkeClusters.%s.label", key)
	name := viper.GetString(label)

	kubecfg := &Kubeconfig{
		Data: data,
		Name: name + "-kubeconfig.yaml",
	}

	return kubecfg
}

func decodeKubeconfig(kubecfg *Kubeconfig) *Kubeconfig {
	enc, ok := kubecfg.Data["kubeconfig"].(string)
	if !ok {
		log.Fatal(errKubeConfig)
	}

	kubeconfig, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		log.Fatal(err)
	}

	kubecfg.Config = string(kubeconfig)

	return kubecfg
}

func writeKubeconfig(kubecfg *Kubeconfig) *Kubeconfig {
	home := os.Getenv("HOME")
	dir := home + "/.kube/config"

	// Wrapping dir in filepath.Clean() sanitizes path from directory traversal.
	if err := os.MkdirAll(filepath.Clean(dir), 0o750); err != nil {
		log.Printf("non-fatal error: %v", err)
	}

	path := filepath.Join(dir, kubecfg.Name)

	err := os.WriteFile(filepath.Clean(path), []byte(kubecfg.Config), 0o600)
	if err != nil {
		log.Printf("non-fatal error: %v", err)
	}

	return kubecfg
}

// helmTemplate is a generic function constrained to types that satisfy the
// HelmConf interface. It receives inputs passed from a HelmTemplate
// receiver method, execute those values against a Go template and returns the
// resulting string. This is a private function to keep config logic within its
// package scope.
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
	tplPath := filepath.Join(projRoot, "cmd/templates/helm", tpl)

	funcMap := template.FuncMap{
		"randInitPass": RandInitPass,
	}

	// Execute template against the final data map
	t := template.Must(template.New(tpl).Funcs(funcMap).ParseFiles(tplPath))

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, &data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// unmarshalSubkey is a generic function constrained to types that satisfy the
// Conf interface. It receives inputs passed from a Get receiver method to
// locate and unmarshal nested YAML config data from a nested key.
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
		return errors.New("viper.GetStringMap error: no data (empty map)")
		// return fmt.Errorf("viper.GetStringMap error: no data (empty map)")
	}

	subKey := fmt.Sprintf("%s.%s", section, key)
	subViper := viper.Sub(subKey)

	if err := subViper.Unmarshal(&cfg); err != nil {
		return err
	}

	return nil
}
