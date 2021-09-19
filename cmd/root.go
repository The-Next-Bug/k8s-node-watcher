package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	selfConfig "The-Next-Bug/k8s-node-watcher/internal/config"
	"The-Next-Bug/k8s-node-watcher/internal/haproxy"
	"The-Next-Bug/k8s-node-watcher/internal/k8s"
	"The-Next-Bug/k8s-node-watcher/internal/mapper"
)

var cfgFile string
var verbosity int

var rootCmd = &cobra.Command{
	Use:   "k8s-node-watcher",
	Short: "A tool to automatically reconfigure HAProxy from k8s nodes.",
	Long: `k8s-nod-watcher watches for node chagnes in a k8s cluster
and modifies the configuration of an HAProxy instance in real time.
It is designed to function against a specifically configured OSS
HAProxy instance.

WARNING: This is alpha software built for a specific use case.`,
	Run: run,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase logging verbosity")
}

func initConfig() {

	// Setup logger verbosity ala Ansible style
	switch verbosity {
	case 0:
		log.SetLevel(log.FatalLevel)
	case 1:
		log.SetLevel(log.WarnLevel)
	case 2:
		log.SetLevel(log.InfoLevel)
	case 3:
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.TraceLevel)
	}

	selfConfig.InitConfig(cfgFile)
}

func resync(mappers []mapper.BackendMapper, interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			log.Debug("sync all triggered")

			for _, mapper := range mappers {
				err := mapper.SyncAll()
				if err != nil {
					log.WithFields(log.Fields{
						"mapper": mapper,
						"err":    err,
					}).Error("unable to sync backend")
				}
			}
		}
	}()

	return ticker
}

func run(cmd *cobra.Command, args []string) {
	config := selfConfig.GetConfig()

	client, err := k8s.New(config)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("unable to create k8s client")
	}

	haProxyClient, err := haproxy.New(config)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("unable to create HAProxy client")
	}

	backends, err := haProxyClient.GetBackendNames()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("unable to load HAProxy backends")
	}

	backendMappers := make([]mapper.BackendMapper, len(backends))
	for i, backend := range backends {
		backendMapper, err := mapper.New(backend, haProxyClient, config.UseExternal)
		if err != nil {
			log.WithFields(log.Fields{
				"errr": err,
			}).Fatal("unable to build mapper")
		}

		backendMappers[i] = backendMapper
	}

	// Start resync operations
	syncTicker := resync(backendMappers, time.Duration(config.ResyncSeconds)*time.Second)
	defer syncTicker.Stop()

	// We need NodeListeners here.
	nodeListeners := make([]k8s.NodeListener, len(backendMappers))
	for i, m := range backendMappers {
		nodeListeners[i] = m
	}

	if err := client.NodeWatch(nodeListeners); err != nil {
		cobra.CheckErr(err.Error())
	}
}
