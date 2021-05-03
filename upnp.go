package main

import (
	"github.com/meshify-app/meshify/model"
	upnp "gitlab.com/NebulousLabs/go-upnp"
)

func ConfigureUPnP(host model.Host) error {

	if host.Current.UPnP == true {
		router, err := upnp.Discover()

		if err != nil {
			return err
		}

		if host.Current.ListenPort != 0 {
			router.Clear(uint16(host.Current.ListenPort))
			router.Forward(uint16(host.Current.ListenPort), host.Name+" "+host.MeshName)
		}
	}

	return nil
}
