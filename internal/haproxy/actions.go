package haproxy

import (
	log "github.com/sirupsen/logrus"
)

func (c *Client) EnableServer(backend, server, ip string) error {
	runtime := c.client.GetRuntime()

	err := runtime.SetServerAddr(backend, server, ip, 0)
	if err != nil {
		log.WithFields(log.Fields{
			"backend": backend,
			"server":  server,
			"ip":      ip,
			"err":     err,
		}).Error("unable to set server address")
		return err
	}

	if err := runtime.EnableServer(backend, server); err != nil {
		log.WithFields(log.Fields{
			"backend": backend,
			"server":  server,
			"err":     err,
		}).Error("unable to enable server")
		return err
	}

	return nil
}

func (c *Client) DisableServer(backend, server string) error {
	runtime := c.client.GetRuntime()

	err := runtime.DisableServer(backend, server)
	if err != nil {
		log.WithFields(log.Fields{
			"backend": backend,
			"server":  server,
			"err":     err,
		}).Error("unable to enable server")
		return err
	}

	return nil
}
