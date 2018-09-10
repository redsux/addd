package ddns

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/redsux/addd/core"
)

var (
	domain = "."
	serial = 1 + rand.Intn(4294967294) // random DNS SOA serial
)

func noDotDomain() string {
	return strings.TrimLeft(domain, ".")
}

func getSoa() *dns.SOA {
	strSoa := fmt.Sprintf("$ORIGIN %s\n@ SOA ns.%s admin. %d 3600 1800 604800 %d", domain, noDotDomain(), serial, 604800)
	soa, err := dns.NewRR(strSoa)
	if err != nil {
		panic(err)
	}
	return soa.(*dns.SOA)
}

func getNS() *dns.NS {
	strNS := fmt.Sprintf("%s %d IN NS ns.%s", domain, 604800, noDotDomain())
	ns, err := dns.NewRR(strNS)
	if err != nil {
		panic(err)
	}
	return ns.(*dns.NS)
}

func getNsA() ([]dns.RR, error) {
	ips, err := addd.IPs()
	if err != nil {
		return nil, err
	}
	ns := make([]dns.RR, 0)
	for _, ip := range ips {
		strRr := fmt.Sprintf("%s %d IN A %s", "ns."+noDotDomain(), 604800, ip)
		rr, err := dns.NewRR(strRr)
		if err != nil {
			return nil, err
		}
		ns = append(ns, rr)
	}
	return ns, nil
}

func updateRecord(r dns.RR, q *dns.Question) int {
	rec, err := addd.NewRecordFromDNS(r)
	if err != nil {
		addd.Log.ErrorF("[DNS] record issue %v.", err)
		return dns.RcodeServerFailure
	}
	addd.Log.NoticeF("[DNS] Update %v, %v", rec.Name, rec.Type)
	if rec.Name == "ns."+domain {
		addd.Log.Warning("[DNS] try to update NS records")
		return dns.RcodeRefused
	}
	if _, err := addd.GetRecord(rec.Name, rec.Type); err == nil {
		if err := addd.DeleteRecord(rec); err != nil {
			addd.Log.ErrorF("[DNS] impossible to delete %v %v", rec.Name, rec.Type)
			addd.Log.DebugF("[DNS] %v", err)
			return dns.RcodeServerFailure
		}
	}
	if err := addd.StoreRecord(rec); err != nil {
		addd.Log.ErrorF("[DNS] impossible to store %v %v", rec.Name, rec.Type)
		addd.Log.DebugF("[DNS] %v", err)
		return dns.RcodeServerFailure
	}
	return dns.RcodeSuccess
}

func queryRecord(q *dns.Question, m *dns.Msg) int {
	qname := strings.ToLower(q.Name)
	qtype := dns.Type(q.Qtype).String()

	addd.Log.NoticeF("[DNS] Query %v, %v", qname, qtype)
	switch q.Qtype {
	case dns.TypeSOA:
		m.Answer = append(m.Answer, getSoa())
		if ns, err := getNsA(); err == nil {
			m.Extra = append(m.Extra, ns...)
		}
	case dns.TypeNS:
		m.Answer = append(m.Answer, getNS())
	case dns.TypeANY:
		qtype = "A"
		fallthrough
	case dns.TypeA:
		if strings.TrimRight(qname,".") == strings.TrimRight("ns."+domain,".") {
			if ns, err := getNsA(); err == nil {
				m.Answer = append(m.Answer, ns...)
			} else {
				addd.Log.DebugF(err.Error())
			}
			break
		}
		fallthrough
	case dns.TypeAAAA:
		readRR, err := addd.GetRecord(qname, qtype)
		if err != nil {
			return dns.RcodeNameError
		}
		rr, err := readRR.DNSRR()
		if err != nil {
			return dns.RcodeServerFailure
		}
		if rr.Header().Name != qname {
			return dns.RcodeBadName
		}
		m.Answer = append(m.Answer, rr)
	default:
		return dns.RcodeNotImplemented
	}
	return dns.RcodeSuccess
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeSuccess)
	m.Authoritative = true
	m.Compress = false
	m.Answer = make([]dns.RR, 0)
	m.Extra = make([]dns.RR, 0)
	m.Ns = []dns.RR{getSoa()}

	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, question := range r.Question {
			if r := queryRecord(&question, m); r > m.Rcode {
				m.Rcode = r
			}
		}
	case dns.OpcodeUpdate:
		for _, question := range r.Question {
			for _, rr := range r.Ns {
				if r := updateRecord(rr, &question); r > m.Rcode {
					m.Rcode = r
				}
			}
		}
	default:
		m.Rcode = dns.RcodeNotImplemented
	}

	addd.Log.DebugF("[DNS] Nb Anwsers = %v\tRcode = %v",len(m.Answer), dns.RcodeToString[m.Rcode])

	if r.IsTsig() != nil {
		if w.TsigStatus() == nil {
			if rtsig, ok := r.Extra[len(r.Extra)-1].(*dns.TSIG); ok {
				m.SetTsig(rtsig.Hdr.Name, dns.HmacMD5, 300, time.Now().Unix())
			}
		} else {
			addd.Log.WarningF("TSIG Status : %v", w.TsigStatus().Error())
		}
	}

	w.WriteMsg(m)
}

func Serve(root, name, secret string, port int) {
	if !strings.HasSuffix(root, ".") {
		root = root + "."
	}
	if strings.HasPrefix(root, ".") && root != "." {
		root = strings.TrimLeft(root, ".")
	}

	root = strings.ToLower(root)

	if _, ok := dns.IsDomainName(root); ok {
		domain = root

		dns.HandleFunc(root, handleDNSRequest)

		server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
		if name != "" && secret != "" {
			server.TsigSecret = map[string]string{name: secret}
		}

		err := server.ListenAndServe()
		defer server.Shutdown()

		if err != nil {
			addd.Log.Error("Failed to setup the udp server.")
			panic(err.Error())
		}
	} else {
		addd.Log.ErrorF("Root domain %v invalid.", root)
	}
}

func ExtractTSIG(tsig string) (name, secret string) {
	if tsig != "" {
		a := strings.SplitN(tsig, ":", 2)
		name, secret = dns.Fqdn(a[0]), a[1]
	}
	return
}
