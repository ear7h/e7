package main

import (
	"testing"
	"time"
	"github.com/ear7h/e7/client"
	"net/http"
	"github.com/miekg/dns"
	"fmt"
	"github.com/ear7h/e7"
	"strconv"
)

func TestRemote(t *testing.T) {
	*SIBLING = "http://104.131.130.194"
	DNS_ADDR = ":4453"

	go main()

	time.Sleep(1 * time.Second)
	DNS_ADDR = ":53"

	port, err := client.Get()
	if err != nil {
		panic(err)
	}

	err = client.Register("test-service", port)
	if err != nil {
		panic(err)
	}

	go http.ListenAndServe(":" + strconv.FormatInt(int64(port), 10), makePingHandler())


	m := new(dns.Msg)

	m.SetQuestion("ear7h.net.", dns.TypeA)

	r, err := dns.Exchange(m, "104.131.130.194"+DNS_ADDR)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

	m.SetQuestion(e7.Hostname()+".ear7h.net.", dns.TypeA)

	r, err = dns.Exchange(m, "104.131.130.194"+DNS_ADDR)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

	m.SetQuestion("test-service.ear7h.net.", dns.TypeA)

	r, err = dns.Exchange(m, "104.131.130.194"+DNS_ADDR)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

}

func TestRemoteDNS(t *testing.T) {
	m := new(dns.Msg)

	m.SetQuestion("ear7h.net.", dns.TypeA)

	r, err := dns.Exchange(m, "104.131.130.194"+DNS_ADDR)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)
}