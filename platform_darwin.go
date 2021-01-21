package main

import (
	"bytes"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

func GetWireguardPath() string {
	return "/usr/local/etc/wireguard/"
}

func ReloadWireguardConfig(meshName string) error {

	args := []string{"wg-quick", "down", meshName}

	cmd := exec.Command("/usr/local/bin/bash", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
	}

	time.Sleep(1 * time.Second)

	args = []string{"wg-quick", "up", meshName}

	cmd = exec.Command("/usr/local/bin/bash", args...)
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
		return err
	}

	return nil

}
