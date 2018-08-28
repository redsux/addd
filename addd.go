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
    loglevel *string
    db_path  *string
    pid_file *string
    // dns flags
    domain   *string
    tsig     *string
    port     *int
    // api flags
    listen   *string
    token    *string
)
 
func main() { 
    // Parse common flags
    loglevel = flag.String("level", "info", "Loglevel (debug>critical>warning>info)")
    db_path = flag.String("db_path", "./addd.db", "location where db will be stored")
    pid_file = flag.String("pid", "./addd.pid", "pid file location")

    // Parse DNS flags
    domain = flag.String("domain", ".", "Parent domain to serve.")
    port = flag.Int("port", 53, "server port")
    tsig = flag.String("tsig", "", "use MD5 hmac tsig: keyname:base64")
 
    // Parse API flags
    listen = flag.String("api", ":1632", "RestAPI listening string ([ip]:port)")
    token  = flag.String("token", "secret", "RestAPI X-AUTH-TOKEN base64 value")

    flag.Parse()
 
    // Define LogLevel
    addd.SetLoglevel(*loglevel)
    
    // Extract TSIG key:secret
    name, secret := ddns.ExtractTSIG(*tsig)

    // Open db
    if err := addd.OpenDB(*db_path); err != nil {
        addd.Log.Critical("Couldn't create internal DB file.")
        panic(err.Error())
    }
    defer addd.CloseDB()
 
    if err := addd.StorePid(*pid_file); err != nil {
        addd.Log.Critical("Couldn't create pid file")
		panic(err.Error())
    }
 
    // Start DNS server
    go ddns.Serve(".", name, secret, *port)
 
    // Start API server
    go api.Serve(*listen, *token, strings.EqualFold(*loglevel, "DEBUG"))

    // Wait SIGINT/SIGTERM
    addd.WaitSig()
}