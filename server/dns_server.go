package main

import (
	"github.com/miekg/dns"
	"github.com/ear7h/e7"
	"fmt"
	"time"
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
		start := time.Now()
		fmt.Println("got dns message for:", r.Question[0].Name)

		msg := new(dns.Msg)

		msg.SetReply(r)
		msg.Authoritative = true

		q := r.Question[0]

		if q.Qtype == dns.TypeSOA {
			msg.Answer = []dns.RR {
				dns.SOA{
					Hdr: dns.RR_Header{
						Name: "ear7h.net.",
						Rrtype: dns.TypeSOA,
						Class: dns.ClassINET,
						Ttl: uint32(l.Timeout),
					},
					Ns: "ns.ear7h.net.",
					Mbox: "julio.grillo98@gmail.com",
					Serial: uint32(l.Mutations),
					Refresh: uint32(l.Timeout.Seconds()),
					Retry: uint32(l.Timeout.Seconds() / 4),
					Expire: uint32(l.Timeout.Seconds() * 2),
					Minttl: uint32(l.Timeout.Seconds() / 2),
				},
			}

			w.WriteMsg(msg)
			return
		}

		rr, ok := l.Query(q.Name)
		if ok {
			msg.Answer = rr
		}

		w.WriteMsg(msg)
		fmt.Println("end: ", time.Now().Sub(start))
		fmt.Println("responding: ", msg.String())
	}
}
