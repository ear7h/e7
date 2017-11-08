package main

import (
	"flag"
	"github.com/ear7h/e7"
	"io/ioutil"
	"net/http"
	"os"
	"fmt"
	"encoding/json"
	"bytes"
)

// TODO PROD: change to 53 and 54
const DNS_PORT = ":4453"
const LEDGER_PORT = ":4434"

//TODO: make tests
//TODO: password
//TODO: ledger client

var SIBLING = flag.String("sibling", "", "ip address of other node")

func init() {
	flag.Parse()
}

func main() {
	pass := "asd"

	l := e7.NewLedger(pass)

	if root := os.Getenv("EAR7H_ROOT"); root != "" {
		l.RootIP = root
	} else if *SIBLING != "" {
		// if a sibling is given
		res, err := http.Get(*SIBLING + LEDGER_PORT)
		if err != nil {
			panic(err)
		}

		byt, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		src := res.Request.Host
		src = src[:len(src) - len(LEDGER_PORT)]
		fmt.Println("src: ", src)
		l = e7.ParseLedger(pass, src,byt)

		// make self known
		blk := e7.Block{}

		l.SignBlock(&blk)

		byt, err = json.Marshal(blk)
		if err != nil {
			panic(err)
		}

		for _, v := range l.Nodes() {
			http.Post(v + LEDGER_PORT, "text/json", bytes.NewReader(byt))
		}

	} else {
		fmt.Println("WARNING: no root or sibling provided")
	}

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

	panic(<-errc)
}
