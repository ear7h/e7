package main

import (
	"testing"
	"github.com/ear7h/e7/e7c"
	"time"
	"github.com/miekg/dns"
	"fmt"
	"github.com/ear7h/e7"
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

	_, err := e7c.Register("test-service")
	if err != nil {
		panic(err)
	}

	m := new(dns.Msg)
	m.SetQuestion("_test-service._tcp.ear7h.net.", dns.TypeSRV)

	r, err := dns.Exchange(m, "127.0.0.1"+DNS_PORT)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)
}
