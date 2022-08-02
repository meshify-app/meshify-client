package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

var config struct {
	Quiet         bool
	MeshifyHost   string
	HostID        string
	ApiKey        string
	ServiceGroup  string
	ServiceApiKey string
	CheckInterval int64
	tls           tls.Config
	SourceAddress string
	sourceAddr    *net.TCPAddr
	Debug         bool
	init          bool
	loaded        bool
	path          *string
}

type configError struct {
	message string
}

func (err *configError) Error() string {
	return err.message
}

func saveConfig() error {
	log.Info("Saving config")
	if config.path == nil {
		return nil
	}
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(GetDataPath()+*config.path, data, 0600)
}

func reloadConfig() error {
	log.Info("Reloading config")
	if config.path == nil {
		return nil
	}
	data, err := ioutil.ReadFile(GetDataPath() + *config.path)
	if err != nil {
		return err
	}
	json.Unmarshal(data, &config)

	log.Infof("MeshifyHost: %s", config.MeshifyHost)
	log.Infof("HostID: %s", config.HostID)
	log.Infof("ApiKey: %s", config.ApiKey)
	log.Infof("Quiet: %t", config.Quiet)

	return nil
}

func loadConfig() error {

	if config.loaded {
		return nil
	}

	if !config.init {
		config.init = true

		// configure defaults
		config.Debug = false
		config.Quiet = false
		config.CheckInterval = 10
		config.SourceAddress = "0.0.0.0"
		config.tls.MinVersion = tls.VersionTLS10

		// load defaults from environment
		config.MeshifyHost = os.Getenv("MESHIFY_HOST")
		config.HostID = os.Getenv("MESHIFY_HOST_ID")
		config.ApiKey = os.Getenv("MESHIFY_API_KEY")
		config.ServiceGroup = os.Getenv("MESHIFY_SERVICE_GROUP")
		config.ServiceApiKey = os.Getenv("MESHIFY_SERVICE_API_KEY")

		if config.MeshifyHost == "" {
			config.MeshifyHost = "https://my.meshify.app"
		}

		// pick up command line arguments
		config.path = flag.String("C", "meshify-client.config.json", "Path to configuration file")
		MeshifyHost := flag.String("server", "", "Meshify server to connect to")
		HostID := flag.String("hostid", "", "Host ID to use")
		ServiceGroup := flag.String("servicegroup", "", "Service group to use")
		ServiceApiKey := flag.String("serviceapikey", "", "Service API key to use")

		ApiKey := flag.String("apikey", "", "API key to use")
		CheckInterval := flag.Int64("interval", 0, "Time interval between maps.  Default is 10 (seconds)")
		quiet := flag.Bool("quiet", false, "Do not output to stdout (only to syslog)")
		sourceStr := flag.String("source", "", "Source address for http client requests")
		flag.Parse()

		// Open the config file specified

		file, err := os.Open(GetDataPath() + *config.path)
		if err != nil && *MeshifyHost == "" && *HostID == "" && *ApiKey == "" && config.HostID == "" && config.ApiKey == "" {
			return err
		}

		// If we could open the config read it, otherwise go with cmd line args
		if err == nil {
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&config)
			if err != nil {
				return err
			}
		}

		if *quiet {
			config.Quiet = *quiet
		}

		if *MeshifyHost != "" {
			config.MeshifyHost = *MeshifyHost
		}
		if *HostID != "" {
			config.HostID = *HostID
		}
		if *ApiKey != "" {
			config.ApiKey = *ApiKey
		}

		if *ServiceGroup != "" {
			config.ServiceGroup = *ServiceGroup
		}
		if *ServiceApiKey != "" {
			config.ServiceApiKey = *ServiceApiKey
		}

		if config.MeshifyHost == "" {
			return &configError{"A meshify-client.config.json file with a MeshifyHost parameter is required"}
		}

		if *CheckInterval != 0 {
			config.CheckInterval = *CheckInterval
		}

		if *sourceStr != "" {
			config.SourceAddress = *sourceStr
		}

		config.sourceAddr, err = net.ResolveTCPAddr("tcp", config.SourceAddress+":0")
		if err != nil {
			return err
		}
		config.loaded = true
		log.Infof("MeshifyHost: %s", config.MeshifyHost)
		log.Infof("HostID: %s", config.HostID)
		log.Infof("ApiKey: %s", config.ApiKey)
		log.Infof("Quiet: %t", config.Quiet)

	} else {
		file, err := os.Open(GetDataPath() + *config.path)
		if err != nil {
			return err
		}
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		if err != nil {
			return err
		}

		log.Infof("MeshifyHost: %s", config.MeshifyHost)
		log.Infof("HostID: %s", config.HostID)
		log.Infof("ApiKey: %s", config.ApiKey)
		log.Infof("Quiet: %t", config.Quiet)

		config.loaded = true
	}
	return nil
}
