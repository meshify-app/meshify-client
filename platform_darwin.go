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

// Return the platform
func Platform() string {
	return "MacOS"
}

func GetStats(mesh string) (string, error) {
	args := []string{"wg", "show", mesh, "transfer"}

	out, err := exec.Command("/usr/local/bin/bash", args...).Output()
	if err != nil {
		log.Errorf("Error getting statistics: %v (%s)", err, string(out))
		return "", err
	}

	return string(out), nil
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
	// remove the file if it exists
	path := GetWireguardPath() + meshName + ".conf"
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	return err

}

func StartContainer(service model.Service) (string, error) {
	return "", nil
}

func CheckContainer(service model.Service) bool {
	return true
}

func StopContainer(service model.Service) error {
	return nil
}

func InService() (bool, error) {
	return true, nil
}

func RunService(svcName string) {
	DoWork()
	DoServiceWork()

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
