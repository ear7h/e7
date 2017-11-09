package e7

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"os"
	"strings"
	"time"
)

/*
e7 SRV protocol

mapping of service names to public addresses
every address in the Ledger is assumed to be another node
*/

var _HOSTNAME string

func Hostname() string {
	return _HOSTNAME
}

func init() {
	var err error
	_HOSTNAME, err = os.Hostname()
	if err != nil {
		panic(err)
	}
}

// sent by a node, specifies which services it has
// this is the data structure for internode communications
type Block struct {
	// specified by node
	// server should prevent collisions
	NodeID    string    `json:"node_id"`
	Signature string    `json:"signature"`
	Timestamp time.Time `json:"timestamp"`

	// array of services the node is hosting
	Services []string `json:"services"`

	// ip address of the author of block
	// filled in by the receiving server
	// json tag for serving the ledger
	IP string `json:"ip"`
}

// This is the structure for an active record
type Record struct {
	Name   string    `json:"name"`
	Target string    `json:"target"`
	IsNode bool      `json:"is_node"`
	Ts     time.Time `json:"ts"`
}

func (r Record) TTL(to time.Duration) (ret time.Duration) {
	ret = time.Until(r.Ts.Add(to))
	if ret < 0 {
		ret = 0
	}
	return
}

// active ledger, stores the history and provides methods
// for quick access of active records
type Ledger struct {
	// this defaults to the machine's host name
	// it is used to identify local nodes within the Ledger
	NodeId string `json:"node_id"`

	RootIP string `json:"root_ip"`

	SelfIP string `json:"self_ip"`

	// keeps track of the number of changes
	// since the creation of ledger
	Mutations int `json:"mutations"`

	// node hostnames and ip addresses
	// A and AAAA records
	ActiveRecords map[string][]Record `json:"active_records"`

	// notice that timeout is an attribute of the Ledger
	// and therefore an attribute of the whole network
	Timeout time.Duration `json:"timeout"`

	// the network password
	password string
}

func NewLedger(pass string) (l *Ledger) {
	l = &Ledger{
		NodeId:        _HOSTNAME,
		ActiveRecords: map[string][]Record{},
		Timeout:       2 * time.Minute, //TODO: change for prod
		password:      pass,
	}

	// auto clean
	go func() {
		for {
			time.Sleep(l.Timeout)
			l.Clean()
		}
	}()

	return
}

func ParseLedger(pass, selfIP string, byt []byte) *Ledger {
	ret := new(Ledger)

	err := json.Unmarshal(byt, ret)
	if err != nil {
		panic(err)
	}

	ret.NodeId = Hostname()
	ret.password = pass
	ret.SelfIP = selfIP

	return ret
}

func (l *Ledger) Bytes() (byt []byte) {
	byt, _ = json.Marshal(l)
	return
}

// Signs the block as originated from the instance of the ledger
// the only preserved field is the services field
func (l *Ledger) SignBlock(b *Block) {
	// TODO: see if this effects the verification
	// to check, history would only have self signed blocks
	b.NodeID = l.NodeId
	b.Timestamp = time.Now()
	b.IP = l.SelfIP

	str := b.NodeID +
		b.IP +
		l.password +
		b.Timestamp.Format(time.RFC3339Nano) +
		strings.Join(b.Services, "")

	sum := sha512.Sum512([]byte(str))
	b.Signature = base64.StdEncoding.EncodeToString(sum[:])
}

func (l *Ledger) verifyBlock(b Block) (ok bool) {
	fmt.Println("verifying")
	if b.NodeID == "" {
		fmt.Println("verification failed, no NodeID")
		return
	}

	str := b.NodeID +
		b.IP +
		l.password +
		b.Timestamp.Format(time.RFC3339Nano) +
		strings.Join(b.Services, "")

	sum := sha512.Sum512([]byte(str))
	shouldSig := base64.StdEncoding.EncodeToString(sum[:])

	return b.Signature == shouldSig
}

// adds the block to the history and SRV to the map
// silently ignores ill-formatted addresses
func (l *Ledger) AddBlock(b Block) (ok bool) {
	fmt.Println("adding block:\n", b)

	defer func() {
		fmt.Println("added: ", ok)
	}()

	ok = l.verifyBlock(b)
	if !ok {
		return
	}

	// add record for the sender of the block
	nodeName := b.NodeID + ".ear7h.net."
	l.ActiveRecords[nodeName] = []Record{
		{
			Name:   nodeName,
			Target: b.IP,
			IsNode: true,
			Ts:     b.Timestamp,
		},
	}

	// recall block.Records is a map of names to addresses
	for _, v := range b.Services {
		l.ActiveRecords[v+".ear7h.net."] = append(l.ActiveRecords[v+".ear7h.net."], Record{
			Name:   v + ".ear7h.net.",
			Target: nodeName,
			IsNode: false,
			Ts:     b.Timestamp,
		})
	}

	ok = true
	l.Mutations ++
	return
}

// returns a map of names to ip addresses
func (l *Ledger) Nodes() (nodes map[string]string) {
	nodes = map[string]string{}

	for k, v := range l.ActiveRecords {
		for _, el := range v {
			if el.IsNode {
				nodes[k] = el.Target
			}
		}
	}

	return
}

func (l *Ledger) Clean() {
	for k, v := range l.ActiveRecords {
		arr := []Record{}
		for _, el := range v {
			if el.TTL(l.Timeout) != 0 {
				arr = append(arr, el)
			}
		}

		if len(arr) == 0 {
			delete(l.ActiveRecords, k)
		} else {
			l.ActiveRecords[k] = arr
		}

		l.Mutations ++
	}
}

// this returns all the resource records matching the query
func (l *Ledger) Query(name string) (rr []dns.RR, ok bool) {
	fmt.Println("query for:"+name)
	fmt.Println("ear7h.net. == "+name, name == "ear7h.net.")

	rr = *new([]dns.RR)

	if ok = name == "ear7h.net."; ok {
		rr = append(rr, &dns.A{
			Hdr: dns.RR_Header{
				Name:   "ear7h.net.",
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(l.Timeout),
			},
			A: net.ParseIP(l.RootIP),
		}, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   "ear7h.net.",
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    uint32(l.Timeout),
			},
			AAAA: net.ParseIP(l.RootIP),
		})
		return
	}

	ars, ok := l.ActiveRecords[name]
	if !ok {
		return
	}

	for _, v := range ars {
		if v.IsNode {
			rr = append(rr, dns.RR(&dns.A{
				Hdr: dns.RR_Header{
					Name:   v.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(v.TTL(l.Timeout)),
				},
				A: net.ParseIP(v.Target),
			}), dns.RR(&dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   v.Name,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    uint32(v.TTL(l.Timeout)),
				},
				AAAA: net.ParseIP(v.Target),
			}))
			continue
		}

		rr = append(rr, dns.RR(&dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   v.Name,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    uint32(v.TTL(l.Timeout)),
			},
			Target: v.Target,
		}))
	}
	return
}
