package main

import (
	"bytes"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

func GetWireguardPath() string {
	return "c:\\Meshify\\"
}

func ReloadWireguardConfig(meshName string) error {

	args := []string{"/uninstalltunnelservice", meshName}

	cmd := exec.Command("wireguard.exe", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
	}

	time.Sleep(1 * time.Second)

	path, err := os.Getwd()
	if path[len(path)-1] != '\\' {
		path = path + "\\"
	}

	args = []string{"/installtunnelservice", path + meshName + ".conf"}

	cmd = exec.Command("wireguard.exe", args...)
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
		return err
	}

	return nil

}
