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

func TestRemotely(t *testing.T) {
	*SIBLING = "http://104.131.130.194"

	go main()

	time.Sleep(1 * time.Second)

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

	r, err := dns.Exchange(m, "104.131.130.194"+DNS_PORT)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

	m.SetQuestion(e7.Hostname()+".ear7h.net.", dns.TypeA)

	r, err = dns.Exchange(m, "104.131.130.194"+DNS_PORT)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

	m.SetQuestion("test-service.ear7h.net.", dns.TypeA)

	r, err = dns.Exchange(m, "104.131.130.194"+DNS_PORT)
	if err != nil {
		panic(err)
	}

	fmt.Println(r)

}