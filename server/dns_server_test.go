package main

import (
	"testing"
	"time"
	"github.com/miekg/dns"
	"fmt"
	"encoding/json"
	"bytes"
	"net/http"
	"io/ioutil"
	"github.com/ear7h/e7"
)

const _TEST_PASS = "asd"

func TestDNS(t *testing.T) {
	l := e7.NewLedger(_TEST_PASS)

	go serveLedger(l)

	b := e7.Block{
		Timestamp: time.Now(),
		Services: []string{"test-service"},
	}

	l.SignBlock(&b)

	byt, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(byt)
	res, err := http.Post("http://127.0.0.1" +LEDGER_ADDR, "text/json", buf)
	if err != nil {
		panic(err)
	}

	byt, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(byt))

	go serveDNS(l)

	m := new(dns.Msg)
	m.SetQuestion(e7.Hostname()+".ear7h.net.", dns.TypeA)

	r, err := dns.Exchange(m, "127.0.0.1"+DNS_ADDR)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

	m.SetQuestion("test-service"+".ear7h.net.", dns.TypeA)

	r, err = dns.Exchange(m, "127.0.0.1"+DNS_ADDR)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)
}
