package haproxy

import (
	log "github.com/sirupsen/logrus"
)

func (c *Client) GetBackendNames() ([]string, error) {
	if len(c.haProxyConfig.Backends) > 0 {
		log.Debug("using configured backends")
		return c.haProxyConfig.Backends, nil
	}

	log.Debug("checking haproxyconfig for backends")

	configClient := c.client.GetConfiguration()

	_, rawBackends, err := configClient.GetBackends("")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("unable to read backends from configuration")

		return nil, err
	}

	backends := make([]string, 0, 1)
	for _, backend := range rawBackends {
		backends = append(backends, backend.Name)
	}

	return backends, nil
}

func (c *Client) GetServerNames(backend string) ([]string, error) {
	runtimeClient := c.client.GetRuntime()

	rawServers, err := runtimeClient.GetServersState(backend)
	if err != nil {
		log.WithFields(log.Fields{
			"err":     err,
			"backend": backend,
		}).Error("unable to get servers for backed")

		return nil, err
	}

	servers := make([]string, 0, 1)
	for _, server := range rawServers {
		servers = append(servers, server.Name)
	}

	return servers, nil
}

func (c *Client) LogBackends() {
	backends, _ := c.GetBackendNames()

	log.WithFields(log.Fields{
		"backends": backends,
	}).Info("found configuration backends")
}

func (c *Client) LogServers() {
	backends, _ := c.GetBackendNames()

	for _, backend := range backends {
		servers, _ := c.GetServerNames(backend)
		log.WithFields(log.Fields{
			"backend": backend,
			"servers": servers,
		}).Debug("servers")
	}
}
