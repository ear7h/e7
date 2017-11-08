package main

import (
	"testing"
	"net/http"
	"github.com/ear7h/e7/e7c"
	"time"
	"io/ioutil"
	"fmt"
)

func TestProxyServer(t *testing.T) {
	go testMain()

	time.Sleep(1 * time.Second)

	lsn, err := e7c.Register("test-service")
	if err != nil {
		panic(err)
	}

	go http.Serve(lsn, makePingHandler())

	fmt.Println("addr: ", lsn.Addr())

	res ,err := http.Get("http://" + lsn.Addr().String())
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