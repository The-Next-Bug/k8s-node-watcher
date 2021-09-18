package main

import (
	"context"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/core/v1"

	selfConfig "The-Next-Bug/k8s-node-watcher/internal/config"
	"The-Next-Bug/k8s-node-watcher/internal/k8s"
)

func main() {

	config := selfConfig.InitConfig()

	client, err := k8s.New(config)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("unable to create k8s client")
	}

	nodeWatch, err := client.Clientset().CoreV1().Nodes().Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	defer nodeWatch.Stop()

	eventChannel := nodeWatch.ResultChan()

	for {
		event := <-eventChannel

		node := event.Object.(*v1.Node)
		addresses := k8s.NewEndpoint(node.Status.Addresses)

		log.WithFields(log.Fields{
			"type":   event.Type,
			"id":     node.Spec.ProviderID,
			"status": addresses,
		}).Infof("node update")
	}
}
