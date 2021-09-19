package haproxy

import (
	"errors"

	log "github.com/sirupsen/logrus"

	selfConfig "The-Next-Bug/k8s-node-watcher/internal/config"

	client_native "github.com/haproxytech/client-native/v2"
	"github.com/haproxytech/client-native/v2/configuration"
	"github.com/haproxytech/client-native/v2/runtime"
)

type Client struct {
	client        *client_native.HAProxyClient
	haProxyConfig *selfConfig.HAProxyConfig
}

func New(config *selfConfig.Config) (*Client, error) {
	haProxyConfig := &config.HAProxy

	log.WithFields(log.Fields{
		"config": haProxyConfig,
	}).Info("haproxy config")

	// Setup HAProxy config
	confClient := &configuration.Client{}
	confParams := configuration.ClientParams{
		ConfigurationFile:      haProxyConfig.ConfigPath,
		Haproxy:                haProxyConfig.Bin,
		TransactionDir:         haProxyConfig.TransactionDir,
		PersistentTransactions: haProxyConfig.PersistentTransactions,
	}

	if err := confClient.Init(confParams); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warn("unable to init HAProxy config client")

		return nil, err
	}

	// List of possible sockets for the runtime client
	sockets := make(map[int]string)

	if haProxyConfig.Socket != "" {
		sockets[0] = haProxyConfig.Socket
	} else if confClient != nil {
		// If we have a config client, attempt to lookup
		// any admin sockets in the config
		_, globalConf, err := confClient.GetGlobalConfiguration("")
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Warn("unable read HAProxy config")
		}

		runtimeAPIs := globalConf.RuntimeAPIs
		for idx, r := range runtimeAPIs {
			sockets[idx] = *r.Address
		}
	} else {
		errStr := "no HAProxy sockets configured"
		log.Error(errStr)
		return nil, errors.New(errStr)
	}

	log.WithFields(log.Fields{
		"sockets": sockets,
	}).Info("using sockets")

	// Setup the Runtime client
	runtimeClient := &runtime.Client{}
	if err := runtimeClient.InitWithSockets(sockets); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to init HAProxy runtime client")
		return nil, err
	}

	// Init HAProxy client
	client := &client_native.HAProxyClient{}
	if err := client.Init(confClient, runtimeClient); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to init HAProxy native client")
		return nil, err
	}

	return &Client{
		client:        client,
		haProxyConfig: haProxyConfig,
	}, nil
}
