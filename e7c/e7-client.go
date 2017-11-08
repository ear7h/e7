package e7c

import (
	"net"
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"strconv"
)

const _LOCAL_SERVER = "/var/ear7h/e7.sock"

var unixClient = http.Client{
	Transport: &http.Transport{
		Dial: func(_, _ string) (net.Conn, error) {
			return net.Dial("unix", _LOCAL_SERVER)
		},
	},
}

type RegistryRequest struct {
	Name string `json:"name"`
}


// Registers the service on the local ledger and all others
func Register(name string) (l net.Listener, err error) {
	r := RegistryRequest{name}

	byt, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	resp, err := unixClient.Post("http://ignore.sock", "text/json", bytes.NewBuffer(byt))
	if err != nil {
		panic(err)
	}

	byt, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	m := map[string]int{}
	err = json.Unmarshal(byt, &m)
	if err != nil {
		panic(err)
	}

	return net.Listen("tcp", ":"+strconv.FormatInt(int64(m["port"]), 10))
}
