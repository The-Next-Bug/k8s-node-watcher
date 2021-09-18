package main

import (
	"context"

  log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/api/core/v1"

	selfConfig "The-Next-Bug/k8s-node-watcher/internal/config"
	"The-Next-Bug/k8s-node-watcher/internal/k8s"
)

func main() {
	/*var kubeconfig *string

	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	
	flag.Parse()*/

	watchConfig := selfConfig.InitConfig()	

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", watchConfig.KubeconfigPath)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodeWatch, err := clientset.CoreV1().Nodes().Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	defer nodeWatch.Stop()

	eventChannel := nodeWatch.ResultChan()

	for {
		event := <- eventChannel

		node := event.Object.(*v1.Node)
		addresses := k8s.NewEndpoint(node.Status.Addresses)

		log.WithFields(log.Fields{
			"type": event.Type,
			"id": node.Spec.ProviderID,
			"status": addresses,
		}).Infof("node update")
	}
}
