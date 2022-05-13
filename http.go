package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func ReadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func stats_handler(w http.ResponseWriter, req *http.Request) {
	log.Infof("stats_handler")
	body, err := GetStats()

	if err != nil {
		log.Error(err)
	}

	io.WriteString(w, body)
}

func startHTTPd() {
	http.HandleFunc("/stats/", stats_handler)

	log.Infof("Starting web server on %s", ":53280")

	err := http.ListenAndServe(":53280", nil)
	if err != nil {
		log.Error(err)
	}

}
