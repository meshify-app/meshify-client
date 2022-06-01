package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

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

type Metrics struct {
	Send int64
	Recv int64
}

func MakeStats(name string, body string) (string, error) {
	meshes := make(map[string]Metrics, 11)
	lines := strings.Split(body, ("\n"))
	for i := 0; i < len(lines); i++ {
		parts := strings.Fields(lines[i])
		if len(parts) < 3 {
			break
		}
		recv, _ := strconv.ParseInt(parts[1], 10, 0)
		send, _ := strconv.ParseInt(parts[2], 10, 0)

		mesh, found := meshes[name]
		if !found {
			mesh = Metrics{Send: 0, Recv: 0}
		}
		mesh.Send += send
		mesh.Recv += recv
		meshes[name] = mesh
	}
	result, err := json.Marshal(meshes)
	return string(result), err
}

func stats_handler(w http.ResponseWriter, req *http.Request) {
	log.Infof("stats_handler")
	// /stats/
	parts := strings.Split(req.URL.Path, "/")
	mesh := parts[2]
	log.Infof("GetStats(%s)", mesh)
	body, err := GetStats(mesh)
	if err != nil {
		log.Error(err)
	}

	stats, err := MakeStats(mesh, body)
	if err != nil {
		log.Error(err)
	}
	log.Infof("Stats: %s", stats)
	w.Header().Add("Access-Control-Allow-Origin", "*")
	io.WriteString(w, stats)
}

func startHTTPd() {
	http.HandleFunc("/stats/", stats_handler)

	log.Infof("Starting web server on %s", ":53280")

	err := http.ListenAndServe(":53280", nil)
	if err != nil {
		log.Error(err)
	}

}
