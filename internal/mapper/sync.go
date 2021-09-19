package mapper

import (
	log "github.com/sirupsen/logrus"

	"The-Next-Bug/k8s-node-watcher/internal/k8s"
)

func (sm *serverMapping) syncServer(m *Mapper) error {
	// If a server is already assigned, nothing to do.
	if sm.server != "" {
		return nil
	}

	server, err := m.nextServer()
	if err != nil {
		return err
	}

	sm.server = server
	return nil
}

func (sm *serverMapping) getIp(m *Mapper) string {
	ip := sm.endpoint.InternalIP
	if m.useExternal {
		ip = sm.endpoint.ExternalIP
	}

	return ip
}

// Synchronize this server mapping onto the given backend
func (sm *serverMapping) Sync(m *Mapper, endpoint *k8s.Endpoint) error {
	// Replace the endpoint in this object if we're given one.
	if endpoint != nil {
		sm.endpoint = endpoint
	}

	ip := sm.getIp(m)
	if ip == "" {
		log.WithFields(log.Fields{
			"backend": m.backend,
			"mapping": sm,
		}).Debug("endpoint has no ip, disabling")

		return sm.Disable(m)
	}

	// Does this mapping have a server?
	err := sm.syncServer(m)
	if err != nil {
		return err
	}

	if err := m.haProxyClient.EnableServer(m.backend, sm.server, ip); err != nil {
		log.WithFields(log.Fields{
			"backend": m.backend,
			"mapping": sm,
			"err":     err,
		}).Error("unable to enable endpoint")

		// Attempt to disable the server.
		sm.Disable(m)

		return err
	}

	log.WithFields(log.Fields{
		"mapping": sm,
		"backend": m.backend,
	}).Info("synced endpoint -> server mapping")

	return nil
}

func (sm *serverMapping) Disable(m *Mapper) error {
	server := sm.server

	// If we don't have a server, there's nothing to do
	if server == "" {
		return nil
	}

	// Release the server name so another endpoint can use it.
	m.releaseServer(sm.server)
	sm.server = ""

	// TODO: This has the potential to lead to configuration drift.
	err := m.haProxyClient.DisableServer(m.backend, server)
	if err != nil {
		log.WithFields(log.Fields{
			"backend": m.backend,
			"sm":      sm,
			"err":     err,
		}).Error("unable to disable server")

		return err
	}

	return nil
}
