package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/meshify-app/meshify/model"
	util "github.com/meshify-app/meshify/util"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var meshifyHostAPIFmt = "%s/api/v1.0/host/%s/status"
var meshifyHostUpdateAPIFmt = "%s/api/v1.0/host/%s"

// Start the channel that iterates the meshify update function
func StartChannel(c chan []byte) {

	log.Infof("StartChannel Meshify Host %s", config.MeshifyHost)
	etag := ""
	var err error

	for {
		content := <-c
		if content == nil {
			break
		}

		err = GetMeshifyConfig(&etag)
		if err != nil {
			log.Errorf("Error getting meshify config: %v", err)
		}
	}
}

func GetMeshifyConfig(etag *string) error {

	host := config.MeshifyHost
	var client *http.Client

	if strings.HasPrefix(host, "http:") {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	} else {
		// Create a transport like http.DefaultTransport, but with the configured LocalAddr
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 60 * time.Second,
				LocalAddr: config.sourceAddr,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
		client = &http.Client{
			Transport: transport,
		}

	}

	if !config.loaded {
		err := loadConfig()
		if err != nil {
			log.Errorf("Failed to load config.")
		}
	}

	var reqURL string = fmt.Sprintf(meshifyHostAPIFmt, host, config.HostID)
	if !config.Quiet {
		log.Infof("  GET %s", reqURL)
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	if req != nil {
		req.Header.Set("X-API-KEY", config.ApiKey)
		req.Header.Set("User-Agent", "meshify-client/1.0")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("If-None-Match", *etag)
	}
	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode == 304 {
		} else if resp.StatusCode == 401 {
			log.Errorf("Unauthorized - looking for another API key")
			// Read the config and find another API key
			buffer, err := ioutil.ReadFile(GetDataPath() + "meshify.conf")
			if err == nil {
				var conf model.Message
				err = json.Unmarshal(buffer, &conf)
				if err == nil {
					found := false
					for _, mesh := range conf.Config {
						for _, host := range mesh.Hosts {

							if host.HostGroup == config.HostID && host.APIKey != config.ApiKey {
								config.ApiKey = host.APIKey
								found = true
								saveConfig()
								break
							}
						}
					}
					if !found {
						if len(conf.Config) == 1 {
							// Only one mesh and getting a 401, so lets delete that mesh too
							err = os.Remove(GetDataPath() + "meshify.conf")
							if err != nil {
								log.Errorf("Failed to delete meshify.conf")
							}
						}
					}
				}
			}
			// pick up any changes from the agent or manually editing the config file.
			reloadConfig()

		} else if resp.StatusCode != 200 {
			log.Errorf("Response Error Code: %v", resp.StatusCode)
			// time.Sleep(10 * time.Second)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("error reading body %v", err)
			}
			log.Debugf("%s", string(body))

			etag2 := resp.Header.Get("ETag")

			if *etag != etag2 {
				log.Infof("etag = %s  etag2 = %s", *etag, etag2)
				*etag = etag2
			} else {
				log.Infof("etag %s is equal", etag2)
			}
			UpdateMeshifyConfig(body)
		}
	} else {
		log.Errorf("ERROR: %v, continuing", err)
	}
	if resp != nil {
		resp.Body.Close()
	}

	return err
}

func UpdateMeshifyHost(host model.Host) error {

	log.Infof("UPDATING HOST: %v", host)
	server := config.MeshifyHost
	var client *http.Client

	if strings.HasPrefix(server, "http:") {
		client = &http.Client{
			Timeout: time.Second * 10,
		}
	} else {
		// Create a transport like http.DefaultTransport, but with the configured LocalAddr
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 60 * time.Second,
				LocalAddr: config.sourceAddr,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
		client = &http.Client{
			Transport: transport,
		}

	}

	var reqURL string = fmt.Sprintf(meshifyHostUpdateAPIFmt, server, host.Id)
	log.Infof("  PATCH %s", reqURL)
	content, err := json.Marshal(host)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", reqURL, bytes.NewBuffer(content))
	if err != nil {
		return err
	}
	if req != nil {
		req.Header.Set("X-API-KEY", host.APIKey)
		req.Header.Set("User-Agent", "meshify-client/1.0")
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err == nil {
		if resp.StatusCode != 200 {
			log.Errorf("PATCH Error: Response %v", resp.StatusCode)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("error reading body %v", err)
			}
			log.Infof("%s", string(body))
		}
	}

	if resp != nil {
		resp.Body.Close()
	}
	if req != nil {
		req.Body.Close()
	}

	return nil
}

// UpdateMeshifyConfig updates the config from the server
func UpdateMeshifyConfig(body []byte) {

	// If the file doesn't exist create it for the first time
	if _, err := os.Stat(GetDataPath() + "meshify.conf"); os.IsNotExist(err) {
		file, err := os.OpenFile(GetDataPath()+"meshify.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err == nil {
			file.Close()
		}
	}

	file, err := os.Open(GetDataPath() + "meshify.conf")

	if err != nil {
		log.Errorf("Error opening meshify.conf file %v", err)
		return
	}
	conf, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		log.Errorf("Error reading meshify config file: %v", err)
		return
	}

	// compare the body to the current config and make no changes if they are the same
	if bytes.Equal(conf, body) {
		return
	} else {
		log.Info("Config has changed, updating meshify.conf")

		// if we can't read the message, immediately return
		var msg model.Message
		err = json.Unmarshal(body, &msg)
		if err != nil {
			log.Errorf("Error reading message from server")
			return
		}

		file, err := os.OpenFile(GetDataPath()+"meshify.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Errorf("Error opening meshify.conf for write: %v", err)
			return
		}
		_, err = file.Write(body)
		file.Close()
		if err != nil {
			log.Infof("Error writing meshify.conf file: %v", err)
			return
		}

		var oldconf model.Message
		err = json.Unmarshal(conf, &oldconf)
		if err != nil {
			log.Errorf("Error reading message from disk")
		}

		log.Debugf("%v", msg)

		// make a copy of the message since UpdateDNS will alter it.
		var msg2 model.Message
		json.Unmarshal(body, &msg2)
		err = UpdateDNS(msg2)
		if err != nil {
			log.Errorf("Error updating DNS configuration: %v", err)
		}

		// Get our local subnets, called here to avoid duplication
		subnets, err := GetLocalSubnets()
		if err != nil {
			log.Errorf("GetLocalSubnets, err = ", err)
		}

		// first, delete any meshes that are no longer in the conf
		for i := 0; i < len(oldconf.Config); i++ {
			found := false
			for j := 0; j < len(msg.Config); j++ {
				if oldconf.Config[i].MeshName == msg.Config[j].MeshName {
					found = true
					break
				}
			}
			if !found {
				log.Infof("Deleting mesh %v", oldconf.Config[i].MeshName)
				StopWireguard(oldconf.Config[i].MeshName)
				os.Remove(GetDataPath() + oldconf.Config[i].MeshName + ".conf")

				for _, host := range oldconf.Config[i].Hosts {
					if host.HostGroup == config.HostID {
						KeyDelete(host.Current.PublicKey)
						KeySave()
					}
				}
			}
		}

		// handle any other changes
		for i := 0; i < len(msg.Config); i++ {
			index := -1
			for j := 0; j < len(msg.Config[i].Hosts); j++ {
				if msg.Config[i].Hosts[j].HostGroup == config.HostID {
					index = j
					break
				}
			}
			if index == -1 {
				log.Errorf("Error reading message %v", msg)
			} else {
				host := msg.Config[i].Hosts[index]
				msg.Config[i].Hosts = append(msg.Config[i].Hosts[:index], msg.Config[i].Hosts[index+1:]...)

				// Configure UPnP as needed
				go ConfigureUPnP(host)

				// If any of the AllowedIPs contain our subnet, remove that entry
				for k := 0; k < len(msg.Config[i].Hosts); k++ {
					allowed := msg.Config[i].Hosts[k].Current.AllowedIPs
					for l := 0; l < len(allowed); l++ {
						inSubnet := false
						_, s, _ := net.ParseCIDR(allowed[l])
						for _, subnet := range subnets {
							if subnet.Contains(s.IP) {
								inSubnet = true
							}
						}
						if inSubnet {
							msg.Config[i].Hosts[k].Current.AllowedIPs = append(allowed[:l], allowed[l+1:]...)
						}
					}
				}

				// Check to see if we have the private key

				key, found := KeyLookup(host.Current.PublicKey)
				if !found {
					KeyAdd(host.Current.PublicKey, host.Current.PrivateKey)
					err = KeySave()
					if err != nil {
						log.Errorf("Error saving key: %s %s", host.Current.PublicKey, host.Current.PrivateKey)
					}
					key, _ = KeyLookup(host.Current.PublicKey)
				}

				// If the private key is blank create a new one and update the server
				if key == "" {
					// delete the old public key
					KeyDelete(host.Current.PublicKey)
					wg, _ := wgtypes.GeneratePrivateKey()
					host.Current.PrivateKey = wg.String()
					host.Current.PublicKey = wg.PublicKey().String()
					KeyAdd(host.Current.PublicKey, host.Current.PrivateKey)
					KeySave()

					host2 := host
					host2.Current.PrivateKey = ""

					// Update meshify with the new public key
					UpdateMeshifyHost(host2)

				} else {
					host.Current.PrivateKey = key
				}

				text, err := DumpWireguardConfig(&host, &(msg.Config[i].Hosts))
				if err != nil {
					log.Errorf("error on template: %s", err)
				}

				// Check the current file and if it's an exact match, do not bounce the service
				path := GetWireguardPath()

				force := false
				var bits []byte

				file, err := os.Open(path + msg.Config[i].MeshName + ".conf")
				if err != nil {
					log.Errorf("Error opening %s for read: %v", msg.Config[i].MeshName, err)
					force = true
				} else {
					bits, err = ioutil.ReadAll(file)
					file.Close()
					if err != nil {
						log.Errorf("Error reading meshify config file: %v", err)
						force = true
					}
				}

				if !force && bytes.Equal(bits, text) {
					log.Infof("*** SKIPPING %s *** No changes!", msg.Config[i].MeshName)
				} else {
					err = StopWireguard(msg.Config[i].MeshName)
					if err != nil {
						log.Errorf("Error stopping wireguard: %v", err)
					}

					err = util.WriteFile(path+msg.Config[i].MeshName+".conf", text)
					if err != nil {
						log.Errorf("Error writing file %s : %s", path+msg.Config[i].MeshName+".conf", err)
					}

					if !host.Enable {
						// Host was disabled when we stopped wireguard above
						log.Infof("Mesh %s is disabled.  Stopped service if running.", msg.Config[i].MeshName)
						// Stopping the service doesn't seem very reliable, stop it again
						if err = StopWireguard(msg.Config[i].MeshName); err != nil {
							log.Errorf("Error stopping wireguard: %v", err)
						}

					} else {
						err = StartWireguard(msg.Config[i].MeshName)
						if err == nil {
							log.Infof("Started %s", msg.Config[i].MeshName)
							log.Infof("%s Config: %v", msg.Config[i].MeshName, msg.Config[i])
						}
					}
				}

			}
		}
	}

}

func GetLocalSubnets() ([]*net.IPNet, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	subnets := make([]*net.IPNet, 0)

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				subnets = append(subnets, v)
			}
		}
	}
	return subnets, nil
}

func StartBackgroundRefreshService() {

	for {

		file, err := os.Open(GetDataPath() + "meshify.conf")
		if err != nil {
			log.Errorf("Error opening meshify.conf for read: %v", err)
			return
		}
		bytes, err := ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			log.Errorf("Error reading meshify config file: %v", err)
			return
		}
		var msg model.Message
		err = json.Unmarshal(bytes, &msg)
		if err != nil {
			log.Errorf("Error reading message from server")
		}

		log.Debugf("%v", msg)

		// Get our local subnets, called here to avoid duplication
		subnets, err := GetLocalSubnets()
		if err != nil {
			log.Errorf("GetLocalSubnets, err = ", err)
		}

		for i := 0; i < len(msg.Config); i++ {
			index := -1
			for j := 0; j < len(msg.Config[i].Hosts); j++ {
				if msg.Config[i].Hosts[j].HostGroup == config.HostID {
					index = j
					break
				}
			}
			if index == -1 {
				log.Errorf("Error reading message %v", msg)
			} else {
				host := msg.Config[i].Hosts[index]
				msg.Config[i].Hosts = append(msg.Config[i].Hosts[:index], msg.Config[i].Hosts[index+1:]...)

				// Configure UPnP as needed
				go ConfigureUPnP(host)

				// If any of the AllowedIPs contain our subnet, remove that entry
				for k := 0; k < len(msg.Config[i].Hosts); k++ {
					allowed := msg.Config[i].Hosts[k].Current.AllowedIPs
					for l := 0; l < len(allowed); l++ {
						inSubnet := false
						_, s, _ := net.ParseCIDR(allowed[l])
						for _, subnet := range subnets {
							if subnet.Contains(s.IP) {
								inSubnet = true
							}
						}
						if inSubnet {
							msg.Config[i].Hosts[k].Current.AllowedIPs = append(allowed[:l], allowed[l+1:]...)
						}
					}
				}
				// Check to see if we have the private key

				key, found := KeyLookup(host.Current.PublicKey)
				if !found {
					KeyAdd(host.Current.PublicKey, host.Current.PrivateKey)
					err = KeySave()
					if err != nil {
						log.Errorf("Error saving key: %s %s", host.Current.PublicKey, host.Current.PrivateKey)
					}
					key, _ = KeyLookup(host.Current.PublicKey)
				}

				// If the private key is blank create a new one and update the server
				if key == "" {
					// delete the old public key
					KeyDelete(host.Current.PublicKey)
					wg, _ := wgtypes.GeneratePrivateKey()
					host.Current.PrivateKey = wg.String()
					host.Current.PublicKey = wg.PublicKey().String()
					KeyAdd(host.Current.PublicKey, host.Current.PrivateKey)
					KeySave()

					host2 := host
					host2.Current.PrivateKey = ""

					// Update meshify with the new public key
					UpdateMeshifyHost(host2)

				} else {
					host.Current.PrivateKey = key
				}

				text, err := DumpWireguardConfig(&host, &(msg.Config[i].Hosts))
				if err != nil {
					log.Errorf("error on template: %s", err)
				}
				path := GetWireguardPath()
				err = util.WriteFile(path+msg.Config[i].MeshName+".conf", text)
				if err != nil {
					log.Errorf("Error writing file %s : %s", path+msg.Config[i].MeshName+".conf", err)
				}

				if !host.Enable {
					StopWireguard(msg.Config[i].MeshName)
					log.Infof("Mesh %s is disabled.  Stopped service if running.", msg.Config[i].MeshName)
				} else {
					err = StartWireguard(msg.Config[i].MeshName)
					if err == nil {
						log.Infof("Started %s", msg.Config[i].MeshName)
						log.Infof("%s Config: %v", msg.Config[i].MeshName, msg.Config[i])
					}
				}

			}
			StartDNS()
		}
		// Do this startup process every hour.  Keeps UPnP ports active, handles laptop sleeps, etc.
		time.Sleep(60 * time.Minute)
	}
}

// DoWork error handler
func DoWork() {
	var curTs int64

	// recover from any panics coming from below
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok := r.(error)
			if !ok {
				log.Fatalf("Fatal Error: %v", err)
			}
		}
	}()

	go func() {

		// Determine current timestamp (the wallclock time we'll retrieve files using)

		c := make(chan []byte)
		go startHTTPd()
		go StartChannel(c)
		go StartDNS()
		go StartBackgroundRefreshService()

		curTs = calculateCurrentTimestamp()

		t := time.Unix(curTs, 0)
		log.Infof("current timestamp = %v (%s)", curTs, t.UTC())

		for {
			time.Sleep(100 * time.Millisecond)
			ts := time.Now()

			if ts.Unix() >= curTs {

				b := []byte("")

				c <- b

				curTs = calculateCurrentTimestamp()
				curTs += config.CheckInterval
			}

		}
	}()
}

func getStatistics() error {
	return nil
}

func calculateCurrentTimestamp() int64 {

	return time.Now().Unix()

}
