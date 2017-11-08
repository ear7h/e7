package main

import (
	"github.com/ear7h/e7"
)

// TODO PROD: change to 53 and 54
const DNS_PORT = ":4453"
const LEDGER_PORT = ":4434"


//TODO: make tests
//TODO: password
//TODO: ledger client

func main() {
	pass := "asd"

	l := e7.NewLedger(pass)

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