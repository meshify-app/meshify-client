package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
)

// GetWireguardPath finds wireguard location for the given platform
func GetWireguardPath() string {

	path := GetDataPath() + "Wireguard\\"
	/*	path, err := exePath()
		if err != nil {
			path = "c:\\meshify\\"
		}
		if path[len(path)-1] != '\\' {
			path = path + "\\"
		}
	*/
	return path
}

func GetDataPath() string {
	return "C:\\ProgramData\\Meshify\\"
}

// StartWireguard restarts the wireguard tunnel on the given platform
func StartWireguard(meshName string) error {

	time.Sleep(1 * time.Second)

	args := []string{"/installtunnelservice", GetWireguardPath() + meshName + ".conf"}

	var out bytes.Buffer
	cmd := exec.Command("wireguard.exe", args...)
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error starting WireGuard: %v (%s)", err, out.String())
		return err
	}

	return nil

}

// StopWireguard stops the wireguard tunnel on the given platform
func StopWireguard(meshName string) error {

	args := []string{"/uninstalltunnelservice", meshName}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "wireguard.exe", args...)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Start()
	if err != nil {
		log.Errorf("Error stopping WireGuard: %v (%s)", err, out.String())
	}

	err = cmd.Wait()

	return err

}

// Windows Main functions

func InService() (bool, error) {
	inService, err := svc.IsWindowsService()

	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	return inService, err
}

func RunService(svcName string) {
	runService(svcName, false)
}

func ServiceManager(svcName string, cmd string) {
	var err error
	switch cmd {
	case "debug":
		runService(svcName, true)
		return
	case "install":
		err = installService(svcName, "Meshify Agent")
	case "remove":
		err = removeService(svcName)
	case "start":
		err = startService(svcName)
	case "stop":
		err = controlService(svcName, svc.Stop, svc.Stopped)
	case "pause":
		err = controlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		err = controlService(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Infof("failed to %s %s: %v", cmd, svcName, nil)
		os.Exit(0)
	}

}
