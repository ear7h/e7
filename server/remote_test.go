package main

import (
	"testing"
	"time"
	"github.com/ear7h/e7/e7c"
	"net/http"
	"github.com/miekg/dns"
	"github.com/ear7h/e7"
	"fmt"
)

func TestRemotely(t *testing.T) {
	*SIBLING = "http://104.131.130.194"

	go main()

	time.Sleep(1 * time.Second)

	lst, err := e7c.Register("test-service")
	if err != nil {
		panic(err)
	}

	go http.Serve(lst, makePingHandler())

	m := new(dns.Msg)
	m.SetQuestion(e7.Hostname()+".ear7h.net.", dns.TypeA)

	r, err := dns.Exchange(m, "104.131.130.194"+DNS_PORT)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

}