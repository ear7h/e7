/*
this server keeps track of the connection in the local machine
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ear7h/e7"
	"github.com/ear7h/e7/e7c"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
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

	//activeConnections[e7.Hostname()] = PROXY_PORT
}

func registerConnections(l *e7.Ledger) {
	blk := e7.Block{
		Services: make([]string, len(activeConnections)),
		IP:       "self",
	}

	var i = 0
	for k := range activeConnections {
		blk.Services[i] = k
		i++
	}

	l.SignBlock(&blk)

	l.AddBlock(blk)

	fmt.Println("registered: ", string(l.Bytes()))

	byt, err := json.Marshal(blk)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range l.Nodes() {
		fmt.Println("sending to: ", v)
		if v == "self" {
			continue
		}
		_, err := http.Post("http://"+v+LEDGER_PORT, "text/json", bytes.NewReader(byt))
		if err != nil {
			fmt.Println("send err: ", err)
		} else {
			fmt.Println("self reg successfull at: ", v)
		}
	}
}

func registerService(name string, l *e7.Ledger) (port int) {
	for port = range avaliablePorts {
		delete(avaliablePorts, port)
		break
	}

	activeConnections[name] = port

	blk := e7.Block{
		Services: []string{name},
		IP:       "self",
	}

	l.SignBlock(&blk)

	l.AddBlock(blk)

	byt, err := json.Marshal(blk)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range l.Nodes() {
		fmt.Println("sending to: ", v)
		if v == "self" {
			continue
		}
		_, err := http.Post("http://"+v+LEDGER_PORT, "text/json", bytes.NewReader(byt))
		if err != nil {
			fmt.Println("send err: ", err)
		} else {
			fmt.Println("self reg successfull at: ", v)
		}
	}

	return

}

func serveLocal(l *e7.Ledger) error {
	os.Remove(LOCAL_SERVER)

	ltn, err := net.Listen("unix", LOCAL_SERVER)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			go cleanConnections()
			go registerConnections(l)
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
