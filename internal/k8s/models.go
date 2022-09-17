package k8s

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
)

// Stores information about discovered nodes and their IPs
type Endpoint struct {
	ID         string
	HostName   string
	InternalIP string
	ExternalIP string
}

func isIpv4(ip string) bool {
	matched, err := regexp.MatchString(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`, ip)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"ip" : ip,
		}).Error("bad ip or regular experession")
		return false
	}

	return matched
}

// Create a new Endpoint from a k8s NodeAddress block
func NewEndpoint(id string, addresses []v1.NodeAddress) *Endpoint {
	e := &Endpoint{
		ID: id,
	}

	for _, item := range addresses {
		switch item.Type {
		case v1.NodeHostName:
			e.HostName = item.Address
		case v1.NodeInternalIP:
			if isIpv4(item.Address) {
				e.InternalIP = item.Address
      } else {
				log.WithFields(log.Fields{
					"ip": item.Address,
					"id": id,
				}).Info("found IPV6 ignoring")
			}
		case v1.NodeExternalIP:
			e.ExternalIP = item.Address
		default:
			// Drop these on the floor
		}
	}

	return e
}

type NodeListener interface {
	Add(node *Endpoint)
	Modify(node *Endpoint)
	Delete(node *Endpoint)
	Bookmark(node *Endpoint)
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%+v", *e)
}
