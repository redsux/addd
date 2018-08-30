package ddns

import (
    "fmt"
    "math/rand"
    "strconv"
    "strings"
    "time"

    "github.com/redsux/addd/core"
    "github.com/miekg/dns"
)

var (
    domain string = "."
    ips    []string
    serial int = 1 + rand.Intn(4294967294) // random DNS SOA serial
)

func init() {
    if ip, e := addd.ExternalIP(); e == nil {
        ips = []string{ip}
    } else {
        ips = make([]string, 0)
    }
}

func getSoa() *dns.SOA {
    strSoa := fmt.Sprintf("$ORIGIN %s\n@ SOA ns.%s admin. %d 1800 900 0604800 604800", domain, domain, serial)
    soa, _ := dns.NewRR(strSoa)
    return soa.(*dns.SOA)
}

func getNs() (ns []dns.RR) {
    ns = make([]dns.RR, 0)
    for i := range ips {
        strRr := fmt.Sprintf("%s %v IN A %s", "ns." + domain, 86400, ips[i])
        if rr, e := dns.NewRR(strRr); e == nil {
            ns = append(ns, rr)
        }
    }
    return
}

func updateRecord(r dns.RR, q *dns.Question) {
    if rec, err := addd.NewRecordFromDns(r); err == nil {
        if rec.Name != "ns." + domain {
            addd.Log.NoticeF("[DNS] Update %v, %v", rec.Name, rec.Type)
            if _, err := addd.GetRecord(rec.Name, rec.Type); err == nil {
                addd.DeleteRecord(rec)
            }
            if _, err := rec.DnsRR(); err == nil {
                addd.StoreRecord(rec)
            }
        }
    }
}
 
func parseQuery(m *dns.Msg) {
    for _, q := range m.Question {
        qname := strings.ToLower(q.Name)
        qtype := dns.Type(q.Qtype).String()

        addd.Log.NoticeF("[DNS] Query %v, %v", qname, qtype)

        switch q.Qtype {
        case dns.TypeSOA:
            if qname == domain {
                m.Answer = append(m.Answer, getSoa())
                m.Extra = append(m.Extra, getNs()...)
            }
        case dns.TypeANY:
            qtype = "A"
            fallthrough
        case dns.TypeA:
            if qname == "ns." + domain {
                m.Answer = append(m.Answer, getNs()...)
                return
            }
            fallthrough
        case dns.TypeAAAA:
            fillAnswer(qname, qtype, &m.Answer)
        }
    }
}

func fillAnswer(qname, qtype string, ans *[]dns.RR) {
    if read_rr, e := addd.GetRecord(qname, qtype); e == nil {
        if rr, e := read_rr.DnsRR(); e == nil && rr.Header().Name == qname {
            *ans = append(*ans, rr)
        }
    }
}
 
func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
    m := new(dns.Msg)
    m.SetReply(r)
    m.Authoritative = true
    m.Compress = false
    m.Answer = make([]dns.RR, 0)
    m.Extra = make([]dns.RR, 0)
    m.Ns = []dns.RR{getSoa()}

    switch r.Opcode {
    case dns.OpcodeQuery:
        parseQuery(m)
    case dns.OpcodeUpdate:
        for _, question := range r.Question {
            for _, rr := range r.Ns {
                updateRecord(rr, &question)
            }
        }
    default:
        m.SetRcode(r, dns.RcodeNotImplemented)
    }

    if m.Rcode != dns.RcodeNotImplemented && len(m.Answer) == 0 {
        m.SetRcode(r, dns.RcodeNameError)
    }

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
 
func Serve(root, name, secret string, port int, extips []string) {
	if !strings.HasSuffix(root, ".") {
		root = root + "."
    }
    root = strings.ToLower(root)

	if _, ok := dns.IsDomainName(root); ok {
        domain = root
        if len(extips) > 0 {
            ips = append(ips, extips...)
        }
        addd.Log.Debugf("Ips : %v", ips)

		dns.HandleFunc(root, handleDnsRequest)

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