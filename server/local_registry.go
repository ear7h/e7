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
	"strconv"
	"time"
	"github.com/ear7h/e7/client"
)

const PORT_MIN = 8080
const PORT_MAX = 8090

// port to bool
var availablePorts = map[int]bool{}

// name to port
var activeConnections = map[string]int{}

func init() {
	for i := PORT_MIN; i <= PORT_MAX; i++ {
		availablePorts[i] = true
	}
}

func cleanConnections() {
	for k, v := range activeConnections {
		port := ":" + strconv.FormatInt(int64(v), 10)
		l, err := net.Listen("tcp", port)
		if err != nil {
			availablePorts[v] = true
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

func registerService(name string, port int, l *e7.Ledger) (err error) {
	 if !availablePorts[port] {
		 return fmt.Errorf("port already in use")
	 }

	if _, ok := activeConnections[name]; ok {
		return fmt.Errorf("name %s already taken", name)
	}

	if port < PORT_MIN || port > PORT_MAX {
		return fmt.Errorf("port %d not in range [%d, %d]", port, PORT_MIN, PORT_MAX)
	}

	delete(availablePorts, port)
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
		go func() {
		for {
			go cleanConnections()
			go registerConnections(l)
			time.Sleep((l.Timeout / 10) * 9)
		}
	}()

	return http.ListenAndServe(LOCAL_PORT, makeLocalHandler(l))
}

func makeLocalHandler(l *e7.Ledger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// get an available port
			var port int
			for port = range availablePorts {break}

			res := struct {
				Port int `json:"port"`
			}{port}

			byt, err := json.Marshal(res)
			if err != nil {
				http.Error(w, "could not marshal response", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(byt)
		case http.MethodPost:
			byt, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "couldn't read body", http.StatusInternalServerError)
				return
			}

			regRequest := new(client.RegistryRequest)

			err = json.Unmarshal(byt, regRequest)
			if err != nil {
				http.Error(w, "couldn't parse json", http.StatusBadRequest)
				return
			}

			err = registerService(regRequest.Name, regRequest.Port, l)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			// echo
			w.Write(byt)
		}
	}
}
