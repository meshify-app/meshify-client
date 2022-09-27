package main

import (
	"net"
	"strings"

	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/meshify-app/meshify/model"
	log "github.com/sirupsen/logrus"
)

func ConfigureUPnP(host model.Host) error {

	if host.Current.UPnP {

		log.Infof("***UPNP*** Configuring UPnP for %s", host.Name)
		clients, _, err := internetgateway1.NewWANIPConnection1Clients()

		if err != nil {
			log.Error("Error discovering gateway, upnp likely not supported. %v", err)
			return err
		}
		if len(clients) == 0 {
			log.Error("***UPNP*** No gateway found, upnp likely not supported.")
			return err
		}
		for _, c := range clients {

			if host.Current.ListenPort != 0 && host.Current.Endpoint != "" {
				// get local ip address
				conn, err := net.Dial("udp", "8.8.8.8:53")
				if err != nil {
					log.Error("Impossible to get local ip address")
				} else {
					defer conn.Close()
					localAddr := conn.LocalAddr().(*net.UDPAddr)

					// get the external ip address
					externalIP, err := c.GetExternalIPAddress()
					if err != nil {
						log.Error("Error getting external ip address, %v", err)
					} else {
						log.Infof("***UPNP*** External IP address: %s", externalIP)
						// compare the externalIP to the endpoint
						parts := strings.Split(host.Current.Endpoint, ":")
						if parts[0] != externalIP && externalIP != "0.0.0.0" {
							log.Error("External IP address does not match endpoint")
							// Update the host endpoint at meshify
							host.Current.Endpoint = externalIP + ":" + parts[1]
							UpdateMeshifyHost(host)
						}
					}

					// delete any old port mappings
					err = c.DeletePortMapping("", uint16(host.Current.ListenPort), "UDP")
					if err != nil {
						log.Error("Error deleting port mapping, %v", err)
					}

					log.Infof("***UPNP*** AddPortMapping: %d %s %d %s %s", host.Current.ListenPort, "UDP", host.Current.ListenPort, localAddr.IP.String(), host.Name+" "+host.MeshName)
					// add port mapping
					err = c.AddPortMapping("", uint16(host.Current.ListenPort), "UDP", uint16(host.Current.ListenPort), localAddr.IP.String(), true, host.Name+" "+host.MeshName, 0)
					if err != nil {
						log.Error("Error adding port mapping, %v", err)
					}
				}
			}
		}
	}

	return nil
}
