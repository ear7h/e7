package client

import (
	"net/http"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"fmt"
)

const _LOCAL_PORT = ":4455"

type RegistryRequest struct {
	Name string `json:"name"`
	Port int `json:"port"`
}

// Registers the service on the local ledger and all others
func Register(name string, port int) (err error) {
	r := RegistryRequest{name, port}

	byt, err := json.Marshal(r)
	if err != nil {
		return
	}

	resp, err := http.Post("http://127.0.0.1" + _LOCAL_PORT, "text/json", bytes.NewBuffer(byt))
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("non 200 response %d: %s", resp.StatusCode, resp.Status)
		return
	}

	byt, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

func Get () (port int, err error) {
	res, err := http.Get("http://" + _LOCAL_PORT)
	if err != nil {
		return
	}

	byt, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	stc := struct {
		Port int `json:"port"`
	}{}

	err = json.Unmarshal(byt, &stc)
	if err != nil {
		return
	}

	port = stc.Port
	return
}
