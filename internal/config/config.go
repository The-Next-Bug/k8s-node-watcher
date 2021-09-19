package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	DefaultTransactionDir = "/tmp/haproxy"
)

type HAProxyConfig struct {
	Socket                 string   `mapstructure:"socket"`
	ConfigPath             string   `mapstructure:"config"`
	Bin                    string   `mapstructure:"bin"`
	PersistentTransactions bool     `mapstructure:"persistent_transactions"`
	TransactionDir         string   `mapstructure:"transaction_dir"`
	Backends               []string `mapstucture:"backends"`
}

type Config struct {
	UseExternal    bool          `mapstructure:"use_external"`
	ResyncSeconds  int64         `mapstructure:"resync_seconds"`
	KubeMaster     string        `mapstructure:"kube_master"`
	KubeconfigPath string        `mapstructure:"kubeconfig"`
	HAProxy        HAProxyConfig `mapstructure:"haproxy"`
}

func defaultKubeconfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Warn("unable to find user home directory")
		return ""
	}

	return filepath.Join(home, ".kube", "config")
}

func InitConfig(cfgFile string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/k8s-node-watcher/")
		viper.AddConfigPath("$HOME/.k8s-node-watcher/")
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warnf("no config file found, using defaults")
	} else {
		log.WithFields(log.Fields{
			"file": viper.ConfigFileUsed(),
		}).Infof("found config file")
	}
}

func GetConfig() *Config {
	config := &Config{
		KubeconfigPath: defaultKubeconfig(),
	}

	if err := viper.UnmarshalExact(&config); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("unable to decode config into struct")
	}

	// Handle some field defaulting

	if config.HAProxy.TransactionDir == "" {
		config.HAProxy.TransactionDir = DefaultTransactionDir
	}

	// Do not allow resync faster than every ten seconds
	if config.ResyncSeconds < 10 {
		config.ResyncSeconds = 10

		log.Warn("defaulting resync interval to 10 seconds")
	}

	log.WithFields(log.Fields{
		"config": config,
	}).Debug("using config")

	return config
}
