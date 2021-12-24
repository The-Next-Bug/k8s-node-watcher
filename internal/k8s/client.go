package k8s

import (
	"context"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watchv1 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"k8s.io/api/core/v1"

	"The-Next-Bug/k8s-node-watcher/internal/config"
)

type Client struct {
	clientset *kubernetes.Clientset
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

func (c *Client) NodeWatch(listeners []NodeListener) error {
	nodeWatch, err := c.clientset.CoreV1().
		Nodes().Watch(context.TODO(), metav1.ListOptions{})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to start Node Watcher")

		return err
	}

	defer nodeWatch.Stop()

	eventChannel := nodeWatch.ResultChan()

	for event := range eventChannel {

		// This will need to be revisited at a later date.
		// TODO: Fix error handling
		if event.Type == watchv1.Error {
			log.WithFields(log.Fields{
				"err":  event.Object,
				"type": event.Type,
			}).Error("error event seen, ignoring")
			continue
		}

		node := event.Object.(*v1.Node)
		endpoint := NewEndpoint(node.Spec.ProviderID, node.Status.Addresses)

		log.WithFields(log.Fields{
			"type":   event.Type,
			"id":     node.Spec.ProviderID,
			"status": endpoint,
		}).Debug("node event")

		for _, listener := range listeners {
			switch event.Type {
			case watchv1.Added:
				listener.Add(endpoint)

			case watchv1.Modified:
				listener.Modify(endpoint)

			case watchv1.Deleted:
				listener.Delete(endpoint)

			case watchv1.Bookmark:
				listener.Bookmark(endpoint)

			default:
				log.WithFields(log.Fields{
					"type":  event.Type,
					"error": event.Object,
				}).Error("unknown event type")
			}
		}
	}

	log.WithFields(log.Fields{}).Warn("watcher disconnected")

	return nil
}
