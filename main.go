package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	inService, err := InService()
	if inService {
		RunService(svcName)
		return
	}

	if len(os.Args) > 1 {
		cmd := strings.ToLower(os.Args[1])

		ServiceManager(cmd)

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
