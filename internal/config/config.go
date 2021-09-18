package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)


type Config struct {
	KubeMaster string `yaml:kube_master`
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
    
	if err := viper.ReadInConfig() ; err != nil {
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
		HAProxyPath: "/etc/haproxy/haproxy.config",
	}


  if err := viper.Unmarshal(&config) ; err != nil {
		log.WithFields(log.Fields{
      "err": err,
    }).Fatal("unable to decode config into struct")
	}

  return config
}

