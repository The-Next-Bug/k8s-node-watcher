package k8s

import (
	log "github.com/sirupsen/logrus"

  clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"The-Next-Bug/k8s-node-watcher/internal/config"
)

type Client struct {
	clientset      *kubernetes.Clientset
}

// This is effectiely a copy of clientcmd.BuildConfigFromFlags with different log
// messages.
func buildKubeconfig(config *config.Config) (*restclient.Config, error) {
	if config.KubeconfigPath == "" && config.KubeMaster == "" {
		log.Warning("no kubeconfig or master url specified, using the inClusterConfig")
		kubeconfig, err := restclient.InClusterConfig()
		
		if err == nil {
			return kubeconfig, nil
		}

		log.Warning("error creating inClusterConfig, falling back to default config: ", err)
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.KubeconfigPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: config.KubeMaster}}).ClientConfig()
}

// Instantiate the k8s client library 
func New(config *config.Config) (*Client, error) {
	kubeconfig, err := buildKubeconfig(config)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to get kubeconfig")
		return nil, err
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to create clientset from kubeconfig")
		return nil, err
	}

	return &Client{
		clientset: clientset,
	}, nil
}

// Placeholder
func (c *Client) Clientset() *kubernetes.Clientset {
	return c.clientset	
}

