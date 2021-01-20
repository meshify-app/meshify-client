package main

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func GetWireguardPath() string {
	return "c:\\Meshify\\"
}

func ReloadWireguardConfig(meshName string) error {

	args := []string{"/uninstalltunnelservice", meshName}

	out, err := exec.Command("wireguard.exe", args...).Output()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out)
	}

	path, err := os.Getwd()
	if path[len(path)-1] != '\\' {
		path = path + "\\"
	}

	args = []string{"/installtunnelservice", path + meshName + ".conf"}

	out, err = exec.Command("wireguard.exe", args...).Output()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out)
		return err
	}

	return nil

}
