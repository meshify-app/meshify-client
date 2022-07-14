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
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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

// Output json structures for the stats and key generation

type Metrics struct {
	Send int64
	Recv int64
}

type Key struct {
	Public string
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

// statHandler will return the stats for the requested mesh
func statsHandler(w http.ResponseWriter, req *http.Request) {
	if !config.Quiet {
		log.Infof("statsHandler")
	}
	// /stats/
	parts := strings.Split(req.URL.Path, "/")
	mesh := parts[2]
	if !config.Quiet {
		log.Infof("GetStats(%s)", mesh)
	}

	// GetStats will execute "wg show mesh transfer" and return the output
	body, err := GetStats(mesh)
	if err != nil {
		log.Error(err)
	}

	// which we then make into a json structure and return it
	stats, err := MakeStats(mesh, body)
	if err != nil {
		log.Error(err)
	}
	if !config.Quiet {
		log.Infof("Stats: %s", stats)
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	io.WriteString(w, stats)
}

// keyHandler will generate a new keypair and insert it into the keystore.
// It will then return the public key.  This allows the agent to create a new
// host without compromising the private key.
func keyHandler(w http.ResponseWriter, req *http.Request) {
	log.Infof("keyHandler")
	// /keys/

	// add the headers here to pass preflight checks
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "*")

	switch req.Method {
	case "GET":
		log.Infof("Method: %s", req.Method)
		key := Key{}
		wg, _ := wgtypes.GeneratePrivateKey()
		key.Public = wg.PublicKey().String()
		KeyAdd(key.Public, wg.String())
		KeySave()
		json.NewEncoder(w).Encode(key)

	default:
		log.Infof("Method: %s", req.Method)
		io.WriteString(w, "")
		log.Infof("Unknown method: %s", req.Method)
	}

}

func startHTTPd() {
	http.HandleFunc("/stats/", statsHandler)
	http.HandleFunc("/keys/", keyHandler)

	log.Infof("Starting web server on %s", ":53280")

	err := http.ListenAndServe(":53280", nil)
	if err != nil {
		log.Error(err)
	}

}
