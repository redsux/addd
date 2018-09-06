package addd

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/redsux/habolt"
)

var (
	bdb *habolt.Store
)

func OpenDB(db_path string) (err error) {
	// Open db
	bdb, err = habolt.NewStore(&habolt.Options{
		Path: db_path,
		BoltOptions: &bolt.Options{
			Timeout: 10 * time.Second,
		},
	})
	return
}

func CloseDB() {
	checkBdp()
	bdb.Close()
}

func ListRecords() (rec []Record, err error) {
	checkBdp()
	rec = make([]Record, 0)
	err = bdb.List(&rec)
	return
}

func GetRecord(domain string, rtype string) (rr *Record, err error) {
	checkBdp()
	rr = DefaultRecord()
    key, _ := getKey(domain, rtype)
	err = bdb.Get(key, rr)
    return 
}

func StoreRecord(rr *Record) (err error) {
	checkBdp()
	key, _ := getKey(rr.Name, rr.Type)
	err = bdb.Set(key, rr)
    return
}

func DeleteRecord(rr *Record) (err error) {
	checkBdp()
    key, _ := getKey(rr.Name, rr.Type)
    err = bdb.Delete(key)
    return
}

func checkBdp() {
	if bdb == nil {
		err := fmt.Errorf("Internal database not define.")
		panic(err)
	}
}