package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"net"
	"os"
)

var config struct {
	Quiet          bool
	Meshify_Host   string
	Host_Id        string
	Check_interval int64
	tls            tls.Config
	Source_address string
	source_addr    *net.TCPAddr
	Debug          bool
}

type ConfigError struct {
	message string
}

func (err *ConfigError) Error() string {
	return err.message
}

func loadConfig() error {
	// Get configuration
	config.Debug = false
	config.Quiet = false
	config.Check_interval = 5
	config.Source_address = "0.0.0.0"
	config.tls.MinVersion = tls.VersionTLS10
	config.Meshify_Host = "https://414n.com/"

	configPath := flag.String("C", "meshify-client.config.json", "Path to configuration file")
	Meshify_Host := flag.String("server", "", "Meshify server to connect to")
	check_interval := flag.Int64("interval", 0, "Time interval between maps.  Default is 5 (5 seconds)")
	quiet := flag.Bool("quiet", false, "Do not output to stdout (only to syslog)")
	source_str := flag.String("source", "", "Source address for http client requests")
	flag.Parse()

	// Open the config file specified

	file, err := os.Open(*configPath)
	if err != nil && *Meshify_Host == "" {
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

	if *Meshify_Host != "" {
		config.Meshify_Host = *Meshify_Host
	}

	if config.Meshify_Host == "" {
		return &ConfigError{"A meshify-client.config.json file with a Meshify_Host parameter is required"}
	}

	if *check_interval != 0 {
		config.Check_interval = *check_interval
	}

	if *source_str != "" {
		config.Source_address = *source_str
	}

	config.source_addr, err = net.ResolveTCPAddr("tcp", config.Source_address+":0")
	if err != nil {
		return err
	}

	return nil
}
