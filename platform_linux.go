package main

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func GetWireguardPath() string {
	return "/etc/wireguard/"
}

func GetDataPath() string {
	return "/etc/meshify/"
}

func DisableHost(meshName string) error {
	args := []string{"wg-quick", "down", meshName}

	cmd := exec.Command("/bin/bash", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error stopping mesh %s: %v (%s)", meshName, err, out.String())
	}
	return err
}

func ReloadWireguardConfig(meshName string) error {

	args := []string{"wg-quick", "down", meshName}

	cmd := exec.Command("/bin/bash", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
	}

	time.Sleep(1 * time.Second)

	args = []string{"wg-quick", "up", meshName}

	cmd = exec.Command("/bin/bash", args...)
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
		return err
	}

	return nil

}

func InService() (bool, error) {
	return true, nil
}

func RunService(svcName string) {

	DoWork()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Errorf("%v", sig)
		done <- true
	}()

	<-done

	log.Info("Exiting")

}

func ServiceManager(svcName string, cmd string) {

}
