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
	dbPath string
	// ha flags
	isHa     bool
	haListen string
	haBind   string
	haJoin   string
)

func init() {
	// Parse common flags
	flag.StringVar(&logLevel, "level", "info", "Loglevel (debug>critical>warning>info)")
	flag.StringVar(&pidFile, "pid", "./addd.pid", "pid file location")

	// Parse DNS flags
	flag.StringVar(&dnsDomain, "domain", "local.", "Parent domain to serve.")
	flag.IntVar(&dnsPort, "port", 53, "server port")
	flag.StringVar(&dnsTsig, "tsig", "", "use MD5 hmac tsig: keyname:base64")

	// Parse API flags
	flag.StringVar(&apiListen, "api", ":1632", "RestAPI listening string ([ip]:port)")
	flag.StringVar(&apiToken, "token", "secret", "RestAPI X-AUTH-TOKEN base64 value")

	// Parse DB flags
	flag.StringVar(&dbPath, "db_path", "./addd.db", "location where db will be stored")

	// HA flags
	flag.BoolVar(&isHa, "ha", false, "Start addd in a HA mode")
	flag.StringVar(&haListen, "ha_listen", ":10001", "Listen to [IP]:PORT (use also PORT+1 for RAFT)")
	flag.StringVar(&haBind, "ha_bind", "", "Bind to a specific [IP]:PORT (for NAT Traversal)")
	flag.StringVar(&haJoin, "ha_join", "", "Members addresses 'host:port' split by a comma ','")
}

func main() {
	var (
		output *habolt.HaOutput
		kvStore habolt.Store
		err error
	)

	// Parse arguments
	flag.Parse()

	// Define LogLevel
	addd.SetLoglevel(logLevel)

	// Extract TSIG key:secret
	dnsName, dnsSecret := ddns.ExtractTSIG(dnsTsig)

	output, err = habolt.NewOutputStr(logLevel)
	if err != nil {
		addd.Log.WarningF("Couldn't create LogOutput with %v", logLevel)
	}
	dbOpts := &habolt.Options{
		Path: dbPath,
		LogOutput: output,
		BoltOptions: &bolt.Options{
			Timeout: 10 * time.Second,
		},
	}

	lAddr, _ := habolt.NewListen(haListen, true)
	bAddr, _ := habolt.NewListen(haBind)

	// Init DB
	if isHa {
		var haStore *habolt.HaStore
		if haStore, err = habolt.NewHaStore(lAddr, bAddr, dbOpts); err != nil {
			addd.Log.Critical("Couldn't create our KV store")
			panic(err.Error())
		}
		var peers []string
		if haJoin != "" {
			for _, peer := range strings.Split(haJoin, ",") {
				if tmpAddr, err := habolt.NewListen(peer); err == nil {
					peers = append(peers, tmpAddr.String())
				}
			}
		}
		go haStore.Start(peers...)
		kvStore = haStore
	} else {
		var stStore *habolt.StaticStore
		if stStore, err = habolt.NewStaticStore(dbOpts); err != nil {
			addd.Log.Critical("Couldn't create our KV store")
			panic(err.Error())
		}
		if bAddr != nil {
			stStore.BindTo(bAddr)
		}
		kvStore = stStore
	}

	// Link DNS to DNS & API
	if err = addd.NewDB(kvStore); err != nil {
		panic(err.Error())
	}
	defer addd.CloseDB()

	if err = addd.StorePid(pidFile); err != nil {
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
