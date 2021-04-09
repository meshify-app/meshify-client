package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/svc"

	log "github.com/sirupsen/logrus"
)

func main() {

	path := "meshify.log"
	file, err := os.Open(GetDataPath() + path)
	if err != nil {

	} else {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(file)
		log.SetLevel(log.InfoLevel)
	}

	err = loadConfig()
	if err != nil && len(os.Args) < 1 {
		log.Fatalf("%v", err)
		os.Exit(1)
	}

	const svcName = "meshify"

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	if inService {
		runService(svcName, false)
		return
	}

	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])
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
			log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
		}
		return
	} else {
		log.Infof("Meshify Control Plane Started")

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

}
