package mapper

import (
	log "github.com/sirupsen/logrus"
)

func (m *Mapper) resetServerPool() error {
	// Populate pool of servers
	serverPool, err := m.haProxyClient.GetServerNames(m.backend)
	if err != nil {
		return err
	}

	m.serverPool = serverPool

	// Clear all server mappings, if there are any
	for _, mapping := range m.serverMap {
		mapping.server = ""
	}

	return nil
}

func (m *Mapper) SyncAll() error {
	log.WithFields(log.Fields{
		"backend": m.backend,
	}).Info("sync triggered")

	if err := m.resetServerPool(); err != nil {
		log.WithFields(log.Fields{
			"backend": m.backend,
			"err":     err,
		}).Error("unable to reset the server pool")
		return err
	}

	m.RLock()
	defer m.RUnlock()

	// Remap all endpoints
	for _, mapping := range m.serverMap {
		err := mapping.Sync(m, nil)
		if err != nil {
			log.WithFields(log.Fields{
				"backend": m.backend,
				"err":     err,
				"mapping": mapping,
			}).Error("unable to resync mapping")

			return err
		}
	}

	// Make sure all servers in the pool are disabled
	for _, server := range m.serverPool {
		mapping := &serverMapping{
			server: server,
		}

		err := mapping.Disable(m)
		if err != nil {
			log.WithFields(log.Fields{
				"backend": m.backend,
				"err":     err,
				"server":  server,
			}).Warn("unable to disable pool server")
		}
	}

	return nil
}
