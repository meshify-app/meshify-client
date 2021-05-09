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
)

var meshifyHostAPIFmt = "%s/api/v1.0/host/%s/status"

type uploadError struct {
	respCode int
}

func (e *uploadError) Error() string {
	return fmt.Sprintf("Http Error %d", e.respCode)
}

// StartHTTPClient starts the client polling
func StartHTTPClient(host string, c chan []byte) {
	log.Infof(" %s", host)
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

	for {
		var content []byte
		content = <-c
		var reqURL string = fmt.Sprintf(meshifyHostAPIFmt, host, config.HostID)
		log.Infof("  GET %s", reqURL)
		req, err := http.NewRequest("GET", reqURL, bytes.NewBuffer(content))
		if err != nil {
			return
		}
		req.Header.Set("X-API-KEY", config.ApiKey)
		req.Header.Set("User-Agent", "meshify-client/1.0")
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode != 200 {
				log.Errorf("Response Error Code: %v, sleeping 10 seconds", resp.StatusCode)
				time.Sleep(10 * time.Second)
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Errorf("error reading body %v", err)
				}
				log.Debugf("%s", string(body))
				UpdateMeshifyConfig(body)
			}
		} else {
			log.Errorf("ERROR: %v, sleeping 10 seconds", err)
			time.Sleep(10 * time.Second)
		}
		if resp != nil {
			resp.Body.Close()
		}
		if req != nil {
			req.Body.Close()
		}

	}
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
		log.Errorf("Error opening config file %v", err)
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
		log.Info("No meshify.conf changes requested")
		return
	} else {
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
		var msg model.Message
		err = json.Unmarshal(body, &msg)
		if err != nil {
			log.Errorf("Error reading message from server")
		}

		log.Debugf("%v", msg)

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
				ConfigureUPnP(host)

				text, err := DumpWireguardConfig(&host, &(msg.Config[i].Hosts))
				if err != nil {
					log.Errorf("error on template: %s", err)
				}
				path := GetWireguardPath()
				err = util.WriteFile(path+msg.Config[i].MeshName+".conf", text)

				if host.Enable == false {
					err = DisableHost(msg.Config[i].MeshName)
					log.Infof("Mesh %s is disabled.  Stopped service if running.", msg.Config[i].MeshName)
				} else {
					err = ReloadWireguardConfig(msg.Config[i].MeshName)
					if err == nil {
						log.Infof("meshify.conf reloaded.  New config:\n%s", body)
					}
				}

			}
		}

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
		go StartHTTPClient(config.MeshifyHost, c)
		go StartDNS()

		curTs = calculateCurrentTimestamp()

		t := time.Unix(curTs, 0)
		log.Infof("current timestamp = %v (%s)", curTs, t.UTC())

		for {
			time.Sleep(100 * time.Millisecond)
			ts := time.Now()

			if ts.Unix() >= curTs {

				err := getStatistics()
				if err != nil {
					log.Errorf("getStatistics: %v", err)
				}

				b := []byte("Alan")

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
