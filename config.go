package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"net"
	"os"
)

var config struct {
	Quiet         bool
	MeshifyHost   string
	HostID        string
	ApiKey        string
	CheckInterval int64
	tls           tls.Config
	SourceAddress string
	sourceAddr    *net.TCPAddr
	Debug         bool
}

type configError struct {
	message string
}

func (err *configError) Error() string {
	return err.message
}

func loadConfig() error {
	// Get configuration
	config.Debug = false
	config.Quiet = false
	config.CheckInterval = 5
	config.SourceAddress = "0.0.0.0"
	config.tls.MinVersion = tls.VersionTLS10
	config.MeshifyHost = "https://dev.meshify.app/"

	configPath := flag.String("C", "meshify-client.config.json", "Path to configuration file")
	MeshifyHost := flag.String("server", "", "Meshify server to connect to")
	CheckInterval := flag.Int64("interval", 0, "Time interval between maps.  Default is 5 (5 seconds)")
	quiet := flag.Bool("quiet", false, "Do not output to stdout (only to syslog)")
	sourceStr := flag.String("source", "", "Source address for http client requests")
	flag.Parse()

	// Open the config file specified

	file, err := os.Open(GetDataPath() + *configPath)
	if err != nil && *MeshifyHost == "" {
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

	if *quiet == true {
		config.Quiet = *quiet
	}

	if *MeshifyHost != "" {
		config.MeshifyHost = *MeshifyHost
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

	return nil
}
