package main

import (
	upnp "github.com/meshify-app/go-upnp"
	"github.com/meshify-app/meshify/model"
	log "github.com/sirupsen/logrus"
)

func ConfigureUPnP(host model.Host) error {

	if host.Current.UPnP {
		router, err := upnp.Discover()

		if err != nil {
			log.Error("Error discovering gateway, upnp likely not supported. %v", err)
			return err
		}

		if host.Current.ListenPort != 0 && host.Current.Endpoint != "" {
			router.Clear(uint16(host.Current.ListenPort))
			err = router.ForwardUDP(uint16(host.Current.ListenPort), host.Name+" "+host.MeshName)
			if err != nil {
				log.Errorf("Error configuring port forward for %d on host%s for mesh %s.  %v", host.Current.ListenPort, host.Name, host.MeshName, err)
			}
		}
	}

	return nil
}
