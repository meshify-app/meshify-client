package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/meshify-app/meshify/model"
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

	// remove the file if it exists
	path := GetWireguardPath() + meshName + ".conf"
	if _, err := os.Stat(path); err == nil {
		os.Remove(path)
	}

	return err

}

// docker run -e MESHIFY_HOST_ID=715d2d3d-2eb2-4f06-be90-4e8d679360a5 -e MESHIFY_API_KEY=example -p 40000:40000 meshify-client

func StartContainer(service model.Service) (string, error) {

	port := fmt.Sprintf("%d", service.ServicePort)

	var args = []string{"run", "--rm", "-d", "--cap-add", "NET_ADMIN", "--cap-add", "SYS_MODULE", "-e", "MESHIFY_HOST_ID=" + service.RelayHost.HostGroup, "-e", "MESHIFY_API_KEY=" + service.RelayHost.APIKey, "-p", port + ":" + port + "/udp", "meshify-client"}
	cmd := exec.Command("docker", args...)

	var outerr bytes.Buffer
	var outstd bytes.Buffer
	cmd.Stderr = &outerr
	cmd.Stdout = &outstd

	err := cmd.Run()
	if err != nil {
		log.Errorf("Error starting container: %v (%s)", err, outerr.String())
		return "", err
	}

	service.Status = "Running"
	if outstd.String() != "" && outstd.String() != "\n" {
		service.ContainerId = outstd.String()
		service.ContainerId = strings.TrimSuffix(service.ContainerId, "\n")
	}

	return service.ContainerId, nil
}

// check the status of the container
func CheckContainer(service model.Service) bool {
	// docker container ls -qf id=3f268613a949
	var args = []string{"container", "ls", "qf", "id=" + service.ContainerId}

	cmd := exec.Command("docker", args...)

	var outerr bytes.Buffer
	var outstd bytes.Buffer
	cmd.Stderr = &outerr
	cmd.Stdout = &outstd

	err := cmd.Run()
	if err != nil {
		log.Errorf("Error checking container: %v (%s)", err, outerr.String())
		return false
	}

	if outstd.String() == service.ContainerId {
		return true
	}

	return false
}

// docker stop service.ContainerId
func StopContainer(service model.Service) error {

	var args = []string{"kill", service.ContainerId}
	cmd := exec.Command("docker", args...)

	var outerr bytes.Buffer
	var outstd bytes.Buffer
	cmd.Stderr = &outerr
	cmd.Stdout = &outstd

	err := cmd.Run()
	if err != nil {
		log.Errorf("Error killing container: %v (%s)", err, outerr.String())
		return err
	}

	service.ContainerId = ""
	service.Status = "Stopped"

	return nil
}

func InService() (bool, error) {
	return true, nil
}

func RunService(svcName string) {

	DoWork()
	DoServiceWork()

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
