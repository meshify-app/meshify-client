package main

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func ReloadWireguardConfig(meshName string) error {

	args := []string{"down", meshName}

	out, err := exec.Command("wg-quick", args...).Output()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out)
	}

	path, err := os.Getwd()
	if path[len(path)-1] != '/' {
		path = path + "/"
	}

	args = []string{"up", meshName}

	out, err = exec.Command("wg-quick", args...).Output()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out)
		return err
	}

	return nil

}
