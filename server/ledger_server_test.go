package main

import (
	"testing"
	"time"
	"net/http"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"github.com/ear7h/e7"
)

func TestLedgerPost(t *testing.T) {
	l := e7.NewLedger(_TEST_PASS)

	go serveLedger(l)

	b := e7.Block{
		NodeID: "testserver",
		Timestamp: time.Now(),
		Services: []string{"test-service"},
	}

	l.SignBlock(&b)

	byt, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(byt)
	res, err := http.Post("http://127.0.0.1" + LEDGER_PORT, "text/json", buf)
	if err != nil {
		panic(err)
	}

	byt, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(byt))
}

func TestLedgerGet(t *testing.T) {
	l := e7.NewLedger(_TEST_PASS)

	go serveLedger(l)

	b := e7.Block{
		NodeID: "testserver",
		Timestamp: time.Now(),
		Services: []string{"test-service"},
	}

	l.SignBlock(&b)

	byt, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(byt)
	_, err = http.Post("http://127.0.0.1" + LEDGER_PORT, "text/json", buf)
	if err != nil {
		panic(err)
	}

	res, err := http.Get("http://127.0.0.1" + LEDGER_PORT)
	byt, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(byt))
}

func TestLedgerPostFail(t *testing.T) {
	l := e7.NewLedger(_TEST_PASS)

	go serveLedger(l)

	b := e7.Block{
		NodeID: "testserver",
		Timestamp: time.Now(),
		Services: []string{"test-service"},
		Signature: "incorrect",
	}

	byt, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(byt)
	_, err = http.Post("http://127.0.0.1" + LEDGER_PORT, "text/json", buf)
	if err != nil {
		panic(err)
	}

	res, err := http.Get("http://127.0.0.1" + LEDGER_PORT)
	if err != nil {
		panic(err)
	}
	byt, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(byt))
}