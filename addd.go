package main
 
import (
    "flag"
    "strings"

    "github.com/redsux/addd/core"
    "github.com/redsux/addd/ddns"
    "github.com/redsux/addd/api"
)
 
var (
    // common flags
    log_level  string
    pid_file   string
    // dns flags
    dns_domain string
    dns_tsig   string
    dns_port   int
    dns_ips    string
    // api flags
    api_listen string
    api_token  string
    // db flags
    db_path    string
    db_listen  string
    db_join    string
)

func init() {
    // Parse common flags
    flag.StringVar(&log_level, "level", "info", "Loglevel (debug>critical>warning>info)")
    flag.StringVar(&pid_file, "pid", "./addd.pid", "pid file location")

    // Parse DNS flags
    flag.StringVar(&dns_domain, "domain", ".", "Parent domain to serve.")
    flag.IntVar(&dns_port, "port", 53, "server port")
    flag.StringVar(&dns_tsig, "tsig", "", "use MD5 hmac tsig: keyname:base64")
    flag.StringVar(&dns_ips, "ips", "", "Add external ips for NS entry (split by ',')")
 
    // Parse API flags
    flag.StringVar(&api_listen, "api", ":1632", "RestAPI listening string ([ip]:port)")
    flag.StringVar(&api_token, "token", "secret", "RestAPI X-AUTH-TOKEN base64 value")

    // Parse DB flags
    flag.StringVar(&db_path, "db_path", "./addd.db", "location where db will be stored")
    //flag.StringVar(&db_listen, "db_port", ":10001", "")
    //flag.StringVar(&db_join, "db_join", "./addd.db", "")
}

func main() { 
    // Parse arguments
    flag.Parse()
 
    // Define LogLevel
    addd.SetLoglevel(log_level)
    
    // Extract TSIG key:secret
    dns_name, dns_secret := ddns.ExtractTSIG(dns_tsig)

    // Open db
    if err := addd.OpenDB(db_path); err != nil {
        addd.Log.Critical("Couldn't create internal DB file.")
        panic(err.Error())
    }
    defer addd.CloseDB()
 
    if err := addd.StorePid(pid_file); err != nil {
        addd.Log.Critical("Couldn't create pid file")
		panic(err.Error())
    }

    // Start DNS server
    go ddns.Serve(dns_domain, dns_name, dns_secret, dns_port, dns_ips)
 
    // Start API server
    go api.Serve(api_listen, api_token, strings.EqualFold(log_level, "DEBUG"))

    // Wait SIGINT/SIGTERM
    addd.WaitSig()
}