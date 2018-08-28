package addd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"github.com/miekg/dns"
)

type Record struct {
	Name    string	`json:"fqdn"    binding:"required"`
	Address string	`json:"address" binding:"required"`
	Type    string	`json:"type"`
	Class   string	`json:"class"`
	Ttl     int		`json:"ttl"`
}

func DefaultRecord() *Record {
	return &Record{
		Type: "A",
		Class: "IN",
		Ttl: 86400,
	}
}

func NewRecord(entry string) ( rec *Record, err error) {
	parsed := strings.Fields(entry)
	if parsed[3] == "A" || parsed[3] == "AAAA" {
		ttl, err := strconv.Atoi(parsed[1])
		if err == nil {
			rec = &Record{
				Name: parsed[0],
				Address: parsed[4],
				Type: parsed[3],
				Class: parsed[2],
				Ttl: ttl,
			}
		}
	} else {
		err = fmt.Errorf("DNS type %v not supported.", parsed[3])
	}
	return
}

func NewRecordFromJson(jso string) ( rec *Record, err error) {
	rec = DefaultRecord()
	err = json.Unmarshal([]byte(jso), rec)
	return
}

func NewRecordFromDns(rr dns.RR) ( rec *Record, err error) {
	var (
		rname string = rr.Header().Name
		rclass string = dns.Class(rr.Header().Class).String()
		rtype string = dns.Type(rr.Header().Rrtype).String()
		rttl int = int(rr.Header().Ttl)
	)
	if _, ok := dns.IsDomainName(rname); ok {
		if a, ok := rr.(*dns.A); ok {
			rec = &Record{
				Name: rname,
				Address: a.A.String(),
				Class: rclass,
				Type: rtype,
				Ttl: rttl,
			}
		} else if a, ok := rr.(*dns.AAAA); ok {
			rec = &Record{
				Name: rname,
				Address: a.AAAA.String(),
				Class: rclass,
				Type: rtype,
				Ttl: rttl,
			}
		} else {
			err = fmt.Errorf("Record %v with type %v not supported.", rname, rtype)
		}
	} else {
		err = fmt.Errorf("Record %v has not a valid domain.", rname)
	}
	return
}

func (r Record) String() string {
	return fmt.Sprintf("%s %v %s %s %s", r.Name, r.Ttl, r.Class, r.Type, r.Address)
}

func (r Record) DnsRR() (rr dns.RR, err error) {
	rr, err = dns.NewRR(r.String())
	return
}

func (r Record) Json() (string,error) {
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
        reverse_domain := strings.Join(labels, ".")
        r = strings.Join([]string{reverse_domain, rtype}, "_")
    } else {
        e = fmt.Errorf("Invalid domain: %v", domain)
    }
    return r, e
}