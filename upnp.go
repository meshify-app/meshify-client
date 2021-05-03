package main

import (
	upnp "github.com/meshify-app/go-upnp"
	"github.com/meshify-app/meshify/model"
)

func ConfigureUPnP(host model.Host) error {

	if host.Current.UPnP == true {
		router, err := upnp.Discover()

		if err != nil {
			return err
		}

		if host.Current.ListenPort != 0 {
			router.Clear(uint16(host.Current.ListenPort))
			router.ForwardUDP(uint16(host.Current.ListenPort), host.Name+" "+host.MeshName)
		}
	}

	return nil
}
