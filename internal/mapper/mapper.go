package mapper

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"The-Next-Bug/k8s-node-watcher/internal/haproxy"
	"The-Next-Bug/k8s-node-watcher/internal/k8s"
)

func New(backend string, haProxyClient *haproxy.Client, useExternal bool) (BackendMapper, error) {

	mapper := &Mapper{
		haProxyClient: haProxyClient,
		useExternal:   useExternal,
		backend:       backend,
		serverMap:     make(map[string]*serverMapping),
	}

	// If we can't get the initial serverPool, something
	// is very broken.
	if err := mapper.resetServerPool(); err != nil {
		return nil, err
	}

	// Now we need to build out a server stream
	return mapper, nil
}

func (m *Mapper) getServer() (string, error) {
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

func (m* Mapper) nextServer() (string, error) {
  next, err  := m.getServer()
  if err == nil {
    return next, nil 
  }

  // Attempt to reset the server pool since we don't seem to have any
  if len(m.serverMap) == 0 {
    log.WithFields(log.Fields{
      "backend": m.backend,
    }).Info("no servers configured, attempting to reset server pool")

    if err := m.resetServerPool(); err != nil {
      log.WithFields(log.Fields{
        "backend": m.backend,
      }).Warn("unble to reset server pool")

      return "", err
    }
  } 

  // try again.
  return m.getServer()
}

func (m *Mapper) releaseServer(server string) {
	m.serverPool = append(m.serverPool, server)
}

// Logs basic statistics about the state of the Mapper
func (m *Mapper) logStats() {
	log.WithFields(log.Fields{
		"backend":     m.backend,
		"server_pool": m.serverPool,
		"server_map":  m.serverMap,
	}).Info("mapper stats")
}

func (m *Mapper) Add(endpoint *k8s.Endpoint) {
	logEvent("added", endpoint)

	m.Lock()
	if _, ok := m.serverMap[endpoint.ID]; ok {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"backend":  m.backend,
		}).Warn("added an existing endpoint")

		// This should not happen.
		m.Unlock()
		m.Modify(endpoint)
		return
	}
	defer m.Unlock()

	mapping := &serverMapping{}
	m.serverMap[endpoint.ID] = mapping

	err := mapping.Sync(m, endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"mapping": mapping,
			"backend": m.backend,
			"err":     err,
		}).Error("unable to sync endpoint")
	}
}

func (m *Mapper) Delete(endpoint *k8s.Endpoint) {
	logEvent("deleted", endpoint)

	m.Lock()
	defer m.Unlock()
	mapping, ok := m.serverMap[endpoint.ID]
	if !ok {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"backend":  m.backend,
		}).Warn("delete for unknown endpoint")

		return
	}

	err := mapping.Disable(m)
	if err != nil {
		log.WithFields(log.Fields{
			"mapping": mapping,
			"backend": m.backend,
		}).Error("unable to delete endpoint")
	}

	// TODO: This could lead to things getting out of sync
	delete(m.serverMap, endpoint.ID)
}

func (m *Mapper) Modify(endpoint *k8s.Endpoint) {
	logEvent("modified", endpoint)

	m.Lock()
	mapping, ok := m.serverMap[endpoint.ID]
	if !ok {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"backend":  m.backend,
		}).Warn("modify for unknown endpoint")

		// This should not happen
		m.Unlock()
		m.Add(endpoint)
		return
	}
	defer m.Unlock()

	err := mapping.Sync(m, endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"mapping": mapping,
			"backend": m.backend,
			"err":     err,
		}).Error("unable to sync endpoint")
	}
}

func (m *Mapper) Bookmark(endpoint *k8s.Endpoint) {
	logEvent("bookmarked", endpoint)
	// Not sure what this does ..
}

func logEvent(msg string, endpoint *k8s.Endpoint) {
	log.WithFields(log.Fields{
		"endpoint": endpoint,
	}).Debug(msg)
}
