package main

import (
	"testing"
	"net/http"
	"github.com/ear7h/e7/client"
	"time"
	"io/ioutil"
	"fmt"
	"strconv"
)

func TestProxyServer(t *testing.T) {
	go testMain()

	time.Sleep(1 * time.Second)

	port, err := client.Get()
	if err != nil {
		panic(err)
	}

	err = client.Register("test-service", port)
	if err != nil {
		panic(err)
	}

	addr := "127.0.0.1:" + strconv.FormatInt(int64(port), 10)

	go http.ListenAndServe(addr, makePingHandler())

	fmt.Println("addr: ", addr)

	res ,err := http.Get("http://" + addr)
	if err != nil {
		panic(err)
	}

	byt, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(byt))

}


func makePingHandler() http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}
}