package main

import (
	"flag"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/redsux/habolt"

	"github.com/redsux/addd/api"
	"github.com/redsux/addd/core"
	"github.com/redsux/addd/ddns"
)

var (
	// common flags
	logLevel string
	pidFile  string
	// dns flags
	dnsDomain string
	dnsTsig   string
	dnsPort   int
	// api flags
	apiListen string
	apiToken  string
	// db flags
	dbPath   string
	dbListen string
	dbJoin   string
)

func init() {
	// Parse common flags
	flag.StringVar(&logLevel, "level", "info", "Loglevel (debug>critical>warning>info)")
	flag.StringVar(&pidFile, "pid", "./addd.pid", "pid file location")

	// Parse DNS flags
	flag.StringVar(&dnsDomain, "domain", ".", "Parent domain to serve.")
	flag.IntVar(&dnsPort, "port", 53, "server port")
	flag.StringVar(&dnsTsig, "tsig", "", "use MD5 hmac tsig: keyname:base64")

	// Parse API flags
	flag.StringVar(&apiListen, "api", ":1632", "RestAPI listening string ([ip]:port)")
	flag.StringVar(&apiToken, "token", "secret", "RestAPI X-AUTH-TOKEN base64 value")

	// Parse DB flags
	flag.StringVar(&dbPath, "dbPath", "./addd.db", "location where db will be stored")
	//flag.StringVar(&dbListen, "db_port", ":10001", "")
	//flag.StringVar(&dbJoin, "dbJoin", "./addd.db", "")
}

func main() {
	// Parse arguments
	flag.Parse()

	// Define LogLevel
	addd.SetLoglevel(logLevel)

	// Extract TSIG key:secret
	dnsName, dnsSecret := ddns.ExtractTSIG(dnsTsig)

	// Init DB
	kvStore, err := habolt.NewStaticStore(&habolt.Options{
		Path: dbPath,
		BoltOptions: &bolt.Options{
			Timeout: 10 * time.Second,
		},
	})
	if err != nil {
		addd.Log.Critical("Couldn't create our KV store")
		panic(err.Error())
	}

	// Open db
	if err := addd.NewDB(kvStore); err != nil {
		panic(err.Error())
	}
	defer addd.CloseDB()

	if err := addd.StorePid(pidFile); err != nil {
		addd.Log.Critical("Couldn't create pid file")
		panic(err.Error())
	}

	// Start DNS server
	go ddns.Serve(dnsDomain, dnsName, dnsSecret, dnsPort)

	// Start API server
	go api.Serve(apiListen, apiToken, strings.EqualFold(logLevel, "DEBUG"))

	// Wait SIGINT/SIGTERM
	addd.WaitSig()
}
