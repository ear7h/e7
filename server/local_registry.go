/*
this server keeps track of the connection in the local machine
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ear7h/e7"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
	"github.com/ear7h/e7/e7c"
)

const LOCAL_SERVER = "/var/ear7h/e7.sock"
const PORT_MIN = 8080
const PORT_MAX = 8090

// port to bool
var avaliablePorts = map[int]bool{}

// name to port
var activeConnections = map[string]int{}

func init() {
	for i := PORT_MIN; i <= PORT_MAX; i++ {
		avaliablePorts[i] = true
	}
}

func cleanConnections() {
	for k, v := range activeConnections {
		port := ":" + strconv.FormatInt(int64(v), 10)
		l, err := net.Listen("tcp", port)
		if err != nil {
			avaliablePorts[v] = true
			delete(activeConnections, k)
		} else {
			l.Close()
		}
	}

	activeConnections[e7.Hostname()] = PROXY_PORT
}

func registerConnections(l *e7.Ledger) {
	blk := e7.Block{
		Services: make([]string, len(activeConnections)),
	}

	// unlike this package which uses port
	// as a key for a service,
	// e7 block uses name as key
	var i = 0
	for k := range activeConnections {
		blk.Services[i] = k
		i++
	}

	l.SignBlock(&blk)

	l.AddBlock(blk)

	byt, err := json.Marshal(blk)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range l.Nodes() {
		http.Post(v+":"+LEDGER_PORT, "text/json", bytes.NewReader(byt))
	}
}

func registerService(name string, l *e7.Ledger) (port int) {
	for port = range avaliablePorts {
		delete(avaliablePorts, port)
		break
	}

	activeConnections[name] = port

	blk := e7.Block{
		Timestamp: time.Now(),
		Services:  []string{name},
	}

	l.SignBlock(&blk)

	l.AddBlock(blk)

	byt, err := json.Marshal(blk)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range l.Nodes() {
		http.Post(v+":"+LEDGER_PORT, "text/json", bytes.NewReader(byt))
	}

	return

}

func serveLocal(l *e7.Ledger) error {
	err := os.Remove(LOCAL_SERVER)
	if err != nil {
		panic(err)
	}

	ltn, err := net.Listen("unix", LOCAL_SERVER)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			cleanConnections()
			registerConnections(l)
			time.Sleep((l.Timeout / 10) * 9)
		}
	}()

	defer ltn.Close()

	return http.Serve(ltn, makeLocalHandler(l))
}

func makeLocalHandler(l *e7.Ledger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			byt, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "couldn't read body", http.StatusInternalServerError)
				return
			}

			regRequest := new(e7c.RegistryRequest)

			err = json.Unmarshal(byt, regRequest)
			if err != nil {
				http.Error(w, "couldn't parse json", http.StatusBadRequest)
				return
			}

			port := registerService(regRequest.Name, l)

			byt, err = json.Marshal(struct {
				Port int `json:"port"`
			}{port})
			if err != nil {
				http.Error(w, "couldn't marshal json", http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusOK)
			w.Write(byt)
		}
	}
}
