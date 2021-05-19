package main

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func GetWireguardPath() string {
	return "/usr/local/etc/wireguard/"
}

func GetDataPath() string {
	return "/usr/local/etc/meshify/"
}

func Startireguard(meshName string) error {

	args := []string{"wg-quick", "up", meshName}

	cmd := exec.Command("/usr/local/bin/bash", args...)
	cmd.Stderr = &out
	go func() {
		err := cmd.Run()
	}()

	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
		return err
	}

	return err

}
func StopWireguard(meshName string) error {

	args := []string{"wg-quick", "down", meshName}

	cmd := exec.Command("/usr/local/bin/bash", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error reloading WireGuard: %v (%s)", err, out.String())
	}

	return err

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
