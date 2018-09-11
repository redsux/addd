package addd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// Record represent our DNS entry object with its Name, Address and types ...
type Record struct {
	Name    string `json:"fqdn"    binding:"required"`
	Address string `json:"address" binding:"required"`
	Type    string `json:"type"`
	Class   string `json:"class"`
	TTL     int    `json:"TTL"`
}

// DefaultRecord create a Record with all default values
func DefaultRecord() *Record {
	return &Record{
		Type:  "A",
		Class: "IN",
		TTL:   86400,
	}
}

// NewRecordFromJSON parse the json in argument to fill a "DefaultRecord()" object
func NewRecordFromJSON(jso string) (rec *Record, err error) {
	rec = DefaultRecord()
	err = json.Unmarshal([]byte(jso), rec)
	rec.Name = strings.TrimRight(rec.Name, ".")
	return
}

// NewRecordFromDNS create a new Record from a dns.RR object
func NewRecordFromDNS(rr dns.RR) (*Record, error) {
	var (
		rname  = strings.ToLower(rr.Header().Name)
		rclass = dns.Class(rr.Header().Class).String()
		rtype  = dns.Type(rr.Header().Rrtype).String()
		rTTL   = int(rr.Header().Ttl)
	)
	rec := &Record{
		Name:    strings.TrimRight(rname, "."),
		Class:   rclass,
		Type:    rtype,
		TTL:     rTTL,
	}
	if _, ok := dns.IsDomainName(rname); ok {
		var ipAddr string
		var v6 bool
		switch a := rr.(type) {
		case *dns.A:
			ipAddr = a.A.String()
			v6 = false
		case *dns.AAAA:
			ipAddr = a.AAAA.String()
			v6 = true
		default:
			err := fmt.Errorf("Record %v with type %v not supported", rname, rtype)
			return nil, err
		}
		if err := IsValidIp(ipAddr, v6); err != nil {
			err := fmt.Errorf("Record %v has invalid ip address %s", rname, ipAddr)
			return nil, err
		}
		rec.Address = ipAddr
	} else {
		err := fmt.Errorf("Record %v has not a valid domain", rname)
		return nil, err
	}
	return rec, nil
}

func (r Record) String() string {
	return fmt.Sprintf("%s %v %s %s %s", r.Name, r.TTL, r.Class, r.Type, r.Address)
}

// DNSRR transforms our object in a dns.RR
func (r Record) DNSRR() (rr dns.RR, err error) {
	rr, err = dns.NewRR(r.String())
	return
}

// JSON transforms our object in a json string
func (r Record) JSON() (string, error) {
	jso, err := json.Marshal(r)
	return string(jso), err
}

func getKey(domain string, rtype string) (r string, e error) {
	if n, ok := dns.IsDomainName(domain); ok {
		labels := dns.SplitDomainName(domain)
		last := n - 1
		for i := 0; i < n/2; i++ {
			labels[i], labels[last-i] = labels[last-i], labels[i]
		}
		reverseDomain := strings.Join(labels, ".")
		r = strings.Join([]string{reverseDomain, rtype}, "_")
	} else {
		e = fmt.Errorf("Invalid domain: %v", domain)
	}
	return r, e
}
