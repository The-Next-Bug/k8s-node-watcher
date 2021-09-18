package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)


type Config struct {
	KubeconfigPath string `yaml:kubeconfig`
	HAProxyPath string `yaml:haproxy`
}

func defaultKubeconfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Warn("unable to find user home directory")
		return ""
	}

	return filepath.Join(home, ".kube", "config")
}

func InitConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/k8s-node-watcher/")
	viper.AddConfigPath("$HOME/.k8s-node-watcher")

	config := &Config{
		KubeconfigPath: defaultKubeconfig(),
		HAProxyPath: "/etc/haproxy/haproxy.config",
	}

	if err := viper.ReadInConfig() ; err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warnf("no config file found")
	} else if err := viper.Unmarshal(&config) ; err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return config
}

