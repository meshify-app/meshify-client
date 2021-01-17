package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var meshify_host_api_fmt = "%s/api/v1.0/host/%s"

func getQueryLogs(path string) ([]byte, error) {
	log.Infof("Retrieving query logs from %s", path)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)

	return b, err
}

type uploadError struct {
	respCode int
}

func (e *uploadError) Error() string {
	return fmt.Sprintf("Http Error %d", e.respCode)
}

func StartHttpClient(host string, c chan []byte) {
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
				LocalAddr: config.source_addr,
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
		var reqUrl string = fmt.Sprintf(meshify_host_api_fmt, host, config.Host_Id)
		log.Infof("  GET %s", reqUrl)
		req, err := http.NewRequest("GET", reqUrl, bytes.NewBuffer(content))
		if err != nil {
			return
		}
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
				log.Infof("%s", string(body))
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
	return
}

func UpdateMeshifyConfig(body []byte) {

	// If the file doesn't exist create it for the first time
	if _, err := os.Stat("meshify.conf"); os.IsNotExist(err) {
		file, err := os.OpenFile("meshify.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err == nil {
			file.Close()
		}
	}

	file, err := os.Open("meshify.conf")

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
		file, err := os.OpenFile("meshify.conf", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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

		err = ReloadWireguardConfig()
		if err == nil {
			log.Infof("meshify.conf reloaded.  New config:\n%s", body)
		}
	}

}

func ReloadWireguardConfig() error {

	args := []string{"setconf", "meshify.conf"}

	out, err := exec.Command("wg", args...).Output()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out)
		return err
	}

	return nil

}

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

		// On startup load the current dnsfilter configuration
		//		err := ReloadDnsfilterConfig()
		//		if err == nil {
		//			log.Infof("Successfully loaded existing dnsfilter configuration")
		//		} else {
		//			log.Errorf("Error loading existing dnsfilter configuration: %v", err)
		//		}

		// Determine current timestamp (the wallclock time we'll retrieve files using)

		c := make(chan []byte)
		go StartHttpClient(config.Meshify_Host, c)

		curTs = CalculateCurrentTimestamp()

		t := time.Unix(curTs, 0)
		log.Infof("current timestamp = %v (%s)", curTs, t.UTC())

		curTs += config.Check_interval

		for {
			time.Sleep(100 * time.Millisecond)
			ts := time.Now()

			if ts.Unix() >= curTs {

				b, err := getQueryLogs("")
				if err != nil {
					log.Errorf("getQueryLogs: %v", err)
				}

				c <- b

				curTs = CalculateCurrentTimestamp()
				curTs += config.Check_interval
			}

		}
	}()
}

func CalculateCurrentTimestamp() int64 {

	now := time.Now().Unix()

	if config.Check_interval == 0 {
		return now
	} else {
		return now - (now % config.Check_interval)
	}

}
