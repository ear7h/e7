package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"flag"
	"github.com/ear7h/e7"
)

func ipFromAddr(remoteAddr string) string {
	arr := strings.Split(remoteAddr, ":")
	if len(arr) == 2 {
		return arr[0]
	}

	// assume everything else is localhost
	return "127.0.0.1"
}

func makeLedgerHandler(l *e7.Ledger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			byt, err := json.Marshal(l)
			if err != nil {
				http.Error(w, "error marshaling json", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(byt)
			return
		case http.MethodPost:
			// decode, verify and add
			byt, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "couldn't read response", http.StatusInternalServerError)
				return
			}

			newBlock := e7.Block{}
			err = json.Unmarshal(byt, &newBlock)
			if err != nil {
				http.Error(w, "couldn't unmarshal json", http.StatusBadRequest)
				return
			}

			newBlock.IP = ipFromAddr(r.RemoteAddr)

			ok := l.AddBlock(newBlock)
			if !ok {
				http.Error(w, "couldn't validate block", http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}
}

var SIBLING = flag.String("sibling", "", "ip address of other node")
func init() {
	flag.Parse()
}

func serveLedger(l *e7.Ledger) error {
	// if a sibling is given
	if *SIBLING != "" {
		res, err := http.Get(*SIBLING + LEDGER_PORT)
		if err != nil {
			panic(err)
		}

		byt, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(byt, &l)
		if err != nil {
			panic(err)
		}
	}

	return http.ListenAndServe(LEDGER_PORT, makeLedgerHandler(l))

}
