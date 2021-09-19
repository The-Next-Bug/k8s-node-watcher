package mapper

import (
	"fmt"

	"The-Next-Bug/k8s-node-watcher/internal/haproxy"
	"The-Next-Bug/k8s-node-watcher/internal/k8s"
)

type BackendMapper interface {
	k8s.NodeListener

	SyncAll() error
}

type serverMapping struct {
	server   string
	endpoint *k8s.Endpoint
}

type Mapper struct {
	haProxyClient *haproxy.Client

	// If set to true, uses the external IP the endpoint
	useExternal bool

	backend string

	// Pool of available servers
	serverPool []string

	// Mapping between nodes and servers
	serverMap map[string]*serverMapping
}

func (sm *serverMapping) String() string {
	return fmt.Sprintf("{server:%s endpoint:%+v", sm.server, sm.endpoint)
}
