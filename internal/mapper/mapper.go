package mapper

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"The-Next-Bug/k8s-node-watcher/internal/haproxy"
	"The-Next-Bug/k8s-node-watcher/internal/k8s"
)

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

	// List of endpoints we can't map to servers
	idleEndpoints []*k8s.Endpoint

	// Mapping between nodes and servers
	serverMap map[string]*serverMapping
}

func New(backend string, haProxyClient *haproxy.Client, useExternal bool) (*Mapper, error) {

	// populate pool of servers
	serverPool, err := haProxyClient.GetServerNames(backend)
	if err != nil {
		return nil, err
	}

	mapper := &Mapper{
		haProxyClient: haProxyClient,
		useExternal:   useExternal,
		backend:       backend,
		serverPool:    serverPool,
		serverMap:     make(map[string]*serverMapping),
	}

	// Now we need to build out a server stream
	return mapper, nil
}

func (m *Mapper) nextServer() (string, error) {
	var next string
	if len(m.serverPool) == 0 {
		log.WithFields(log.Fields{
			"backend": m.backend,
		}).Warn("server pool exhausted")
		return "", fmt.Errorf("server pool exhausted")
	}

	next, m.serverPool = m.serverPool[0], m.serverPool[1:]

	return next, nil
}

func (m *Mapper) releaseServer(server string) {
	m.serverPool = append(m.serverPool, server)
}

// Logs basic statistics about the state of the Mapper
func (m *Mapper) logStats() {
	log.WithFields(log.Fields{
		"backend":        m.backend,
		"idle_endpoints": m.idleEndpoints,
		"server_pool":    m.serverPool,
		"server_map":     m.serverMap,
	}).Info("mapper stats")
}

func (m *Mapper) mapServer(endpoint *k8s.Endpoint) (*serverMapping, error) {
	server, err := m.nextServer()
	if err != nil {
		return nil, err
	}

	mapping := &serverMapping{
		server:   server,
		endpoint: endpoint,
	}

	ip := endpoint.InternalIP
	if m.useExternal {
		ip = endpoint.ExternalIP
	}

	if ip == "" {
		return nil, fmt.Errorf("no ip for backend")
	}

	if err := m.haProxyClient.EnableServer(m.backend, server, ip); err != nil {
		return nil, err
	}

	m.serverMap[endpoint.ID] = mapping

	return mapping, nil
}

func (m *Mapper) Add(endpoint *k8s.Endpoint) {
	logEvent("added", endpoint)

	mapping, err := m.mapServer(endpoint)
	if err != nil {
		// If no server is available, mark this endpoint as idle
		m.idleEndpoints = append(m.idleEndpoints, endpoint)

		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"backend":  m.backend,
		}).Warn("endpoint idle, no server available")

		return
	}

	log.WithFields(log.Fields{
		"enpdpoint": endpoint,
		"server":    mapping.server,
		"backend":   m.backend,
	}).Info("adding endpoint -> server mapping")

	m.logStats()
}

func (m *Mapper) Delete(endpoint *k8s.Endpoint) {
	logEvent("deleted", endpoint)

	m.logStats()
}

func (m *Mapper) Modify(endpoint *k8s.Endpoint) {
	logEvent("modified", endpoint)

	m.logStats()
}

func (m *Mapper) Bookmark(endpoint *k8s.Endpoint) {
	logEvent("bookmarked", endpoint)

	m.logStats()
}

func logEvent(msg string, endpoint *k8s.Endpoint) {
	log.WithFields(log.Fields{
		"endpoint": endpoint,
	}).Debug(msg)
}
