package addd

import (
	"errors"
	"fmt"

	"github.com/redsux/habolt"
)

var (
	bdb habolt.Store
)

// NewDB initialize our key/value store
func NewDB(db habolt.Store) error {
	if db == nil {
		return errors.New("NewDB nil argument not allowed")
	}
	bdb = db
	return nil
}

// CloseDB ends the DB usage
func CloseDB() error {
	checkBdp()
	return bdb.Close()
}

// ListRecords returns all Record stored in our DB
func ListRecords() (rec []Record, err error) {
	checkBdp()
	rec = make([]Record, 0)
	err = bdb.List(&rec)
	return
}

// GetRecord retrieves the record
func GetRecord(domain string, rtype string) (rr *Record, err error) {
	checkBdp()
	key, err := getKey(domain, rtype)
	if err != nil {
		return
	}
	rr = DefaultRecord()
	err = bdb.Get(key, rr)
	return
}

// StoreRecord stores the record in our DB
func StoreRecord(rr *Record) (err error) {
	checkBdp()
	key, err := getKey(rr.Name, rr.Type)
	if err != nil {
		return
	}
	err = bdb.Set(key, rr)
	return
}

// DeleteRecord deletes the record in our DB
func DeleteRecord(rr *Record) (err error) {
	checkBdp()
	key, err := getKey(rr.Name, rr.Type)
	if err != nil {
		return
	}
	err = bdb.Delete(key)
	return
}

func checkBdp() {
	if bdb == nil {
		err := fmt.Errorf("Internal database not define")
		panic(err)
	}
}
