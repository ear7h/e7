package main

import (
	"github.com/miekg/dns"
	"github.com/ear7h/e7"
	"fmt"
)

func serveDNS(l *e7.Ledger) error {
	ret := make(chan error, 1)

	dns.HandleFunc(".", makeDNSHandler(l))
	go func() {
		srv := &dns.Server{Addr: DNS_ADDR, Net: "udp"}
		ret <- srv.ListenAndServe()

	}()
	go func() {
		srv := &dns.Server{Addr: DNS_ADDR, Net: "tcp"}
		ret <- srv.ListenAndServe()
	}()

	return <-ret
}

func makeDNSHandler(l *e7.Ledger) dns.HandlerFunc {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		fmt.Println("got dns message, ", r.String())
		
		msg := new(dns.Msg)

		msg.SetReply(r)
		msg.Authoritative = true

		q := r.Question[0]

		rr, ok := l.Query(q.Name)
		if ok {
			msg.Answer = rr
		}

		w.WriteMsg(msg)
	}
}
