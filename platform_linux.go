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
	return "/etc/wireguard/"
}

func GetDataPath() string {
	return "/etc/meshify/"
}

// Return the platform
func Platform() string {
	return "Linux"
}

func GetStats(mesh string) (string, error) {
	args := []string{"show", mesh, "transfer"}

	out, err := exec.Command("wg", args...).Output()
	if err != nil {
		log.Errorf("Error getting statistics: %v (%s)", err, string(out))
		return "", err
	}

	return string(out), nil
}

func StartWireguard(meshName string) error {

	args := []string{"wg-quick", "up", meshName}

	cmd := exec.Command("/bin/bash", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error starting WireGuard: %v (%s)", err, out.String())
		return err
	}

	return err

}

func StopWireguard(meshName string) error {

	args := []string{"wg-quick", "down", meshName}

	cmd := exec.Command("/bin/bash", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error stopping WireGuard: %v (%s)", err, out.String())
	}

	return err

}

func InService() (bool, error) {
	return true, nil
}

func RunService(svcName string) {

	DoWork()

	log.Info("setting up signal handlers")
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
	os.Exit(0)

}

func ServiceManager(svcName string, cmd string) {

}
