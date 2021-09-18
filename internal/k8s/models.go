package k8s

import (
	"k8s.io/api/core/v1"
)

// Stores information about discovered nodes and their IPs
type Endpoint  struct {
	HostName   string
	InternalIP string
	ExternalIP string
}

// Create a new Endpoint from a k8s NodeAddress block
func NewEndpoint(addresses []v1.NodeAddress) *Endpoint {
	e := &Endpoint{}

	for _, item := range addresses {
		switch(item.Type) {
			case v1.NodeHostName:
				e.HostName = item.Address
			case v1.NodeInternalIP:
				e.InternalIP = item.Address
			case v1.NodeExternalIP:
				e.ExternalIP = item.Address
			default:
				// Drop these on the floor
		}
	}

	return e
}


