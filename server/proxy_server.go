/*
This file contains all client facing serving

A reverse proxy for requests to local services
A an index endpoint which replies with all services registered locally
 */
package main

import (
	"github.com/ear7h/e7"
	"net/http"
	"net/http/httputil"
	"fmt"
	"github.com/gin-gonic/gin/json"
)

const INDEX_PORT = ":8079"
const PROXY_PORT = 4443
const LOCALHOST = "127.0.0.1"

func serveProxy(l *e7.Ledger) error {
	go http.ListenAndServe(INDEX_PORT, makeIndexHandler(l))
	return http.ListenAndServe(fmt.Sprintf(":%d", PROXY_PORT), makeProxy(l))
}

func makeIndexHandler(l *e7.Ledger) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		byt, err := json.Marshal(activeConnections)
		if err != nil {
			http.Error(w, "couldn't marshal json", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(byt)
	}
}

func makeProxy(l *e7.Ledger) http.Handler {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Host = LOCALHOST

			port, ok := activeConnections[r.Host]
			if !ok {
				r.URL.Host = fmt.Sprintf("%s:%d", LOCALHOST, INDEX_PORT)
				return
			}
			r.URL.Host = fmt.Sprintf("%s:%d", LOCALHOST, port)

		},
	}
}