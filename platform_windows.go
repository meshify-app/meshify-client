package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

// GetWireguardPath finds wireguard location for the given platform
func GetWireguardPath() string {
	path, err := os.Getwd()
	if err != nil {
		path = "c:\\meshify\\"
	}
	if path[len(path)-1] != '\\' {
		path = path + "\\"
	}

	return path
}

// ReloadWireguardConfig restarts the wireguard service on the given platform
func ReloadWireguardConfig(meshName string) error {

	args := []string{"/uninstalltunnelservice", meshName}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wireguard.exe", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Start()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
	}

	err = cmd.Wait()

	time.Sleep(1 * time.Second)

	args = []string{"/installtunnelservice", GetWireguardPath() + meshName + ".conf"}

	cmd = exec.Command("wireguard.exe", args...)
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
		return err
	}

	return nil

}
