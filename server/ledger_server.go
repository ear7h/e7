package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"github.com/ear7h/e7"
	"fmt"
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
			w.WriteHeader(http.StatusOK)
			w.Write(l.Bytes())
			return
		case http.MethodPost:
			fmt.Println("block recieved")

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

func serveLedger(l *e7.Ledger) error {

	return http.ListenAndServe(LEDGER_ADDR, makeLedgerHandler(l))

}
