package ddns

import (
    "strconv"
    "strings"
    "time"

    "github.com/redsux/addd/core"
    "github.com/miekg/dns"
)
 
func updateRecord(r dns.RR, q *dns.Question) {
    if rec, err := addd.NewRecordFromDns(r); err == nil {
        addd.Log.NoticeF("[DNS] Update %v, %v", rec.Name, rec.Type)
        if _, err := addd.GetRecord(rec.Name, rec.Type); err == nil {
            addd.DeleteRecord(rec)
        }
        if _, err := rec.DnsRR(); err == nil {
            addd.StoreRecord(rec)
        }
    }
}
 
func parseQuery(m *dns.Msg) {
    var rr dns.RR
    for _, q := range m.Question {
        qtype := dns.Type(q.Qtype).String()
        
        addd.Log.NoticeF("[DNS] Query %v, %v", q.Name, qtype)
        if read_rr, e := addd.GetRecord(q.Name, qtype); e == nil {
            if rr, e = read_rr.DnsRR(); e == nil && rr.Header().Name == q.Name {
                m.Answer = append(m.Answer, rr)
            }
        }
    }
}
 
func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
    m := new(dns.Msg)
    m.SetReply(r)
    m.Compress = false
 
    switch r.Opcode {
    case dns.OpcodeQuery:
        parseQuery(m)
 
    case dns.OpcodeUpdate:
        for _, question := range r.Question {
            for _, rr := range r.Ns {
                updateRecord(rr, &question)
            }
        }
    }
 
    if r.IsTsig() != nil {
        if w.TsigStatus() == nil {
            m.SetTsig(r.Extra[len(r.Extra)-1].(*dns.TSIG).Hdr.Name,
                dns.HmacMD5, 300, time.Now().Unix())
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

	if _, ok := dns.IsDomainName(root); ok {
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