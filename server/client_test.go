package main

import (
	"testing"
	"github.com/ear7h/e7/client"
	"time"
	"github.com/miekg/dns"
	"fmt"
	"github.com/ear7h/e7"
	"net/http"
	"strconv"
)

func testMain() {

	l := e7.NewLedger(_TEST_PASS)

	errc := make(chan error, 1)

	go func() {
		errc <- serveLedger(l)
	}()

	go func() {
		errc <- serveDNS(l)
	}()

	go func() {
		errc <- serveLocal(l)
	}()

	go func() {
		errc <- serveProxy(l)
	}()

	panic(<- errc)
}

func TestRegister(t *testing.T) {

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


	m := new(dns.Msg)

	m.SetQuestion("test-service.ear7h.net.", dns.TypeA)

	r, err := dns.Exchange(m, "127.0.0.1"+DNS_PORT)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)
}
