package app

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

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
	*NodePoolConfig | *LkeConfig
}

type ConfGetter[T Conf] interface {
	Get() error
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
	viper.AddConfigPath(".")

	// Get absolute path to project root
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("InitConfig func error: runtime caller failed to get source file path")
	}

	dir := filepath.Dir(filename)
	projRoot := filepath.Join(dir, "..", "..", "..")

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
