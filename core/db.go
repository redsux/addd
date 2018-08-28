package addd

import (
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

var (
	bdb *bolt.DB
)

const (
	rr_bucket = "rr"
)

func OpenDB(db_path string) (err error) {
	// Open db
	bdb, err = bolt.Open(
		db_path,
		0600,
		&bolt.Options{
			Timeout: 10 * time.Second,
		},
	)
	if err == nil {
		// Create bucket if doesn't exist
		err = createBucket(rr_bucket)
	}
	return
}

func CloseDB() {
	checkBdp()
	bdb.Close()
}

func ListRecords() (rec []*Record, err error) {
	checkBdp()
	rec = make([]*Record, 0)
	err = bdb.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(rr_bucket)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if r, e := NewRecordFromJson( string(v) ); e == nil {
				rec = append(rec, r)
			}
		}
		return nil
	})
	return
}

func GetRecord(domain string, rtype string) (rr *Record, err error) {
	checkBdp()
    key, _ := getKey(domain, rtype)
    var v []byte
 
    err = bdb.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(rr_bucket))
        v = b.Get([]byte(key))
 
        if string(v) == "" {
			e := fmt.Errorf("Record not found : %v", domain)
			Log.Debug(e.Error())
 
            return e
        }
 
        return nil
    })
 
    if err == nil {
		rr, err = NewRecordFromJson( string(v) )
    }
 
    return rr, err
}

func StoreRecord(rr *Record) (err error) {
	checkBdp()
    key, _ := getKey(rr.Name, rr.Type)
    err = bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(rr_bucket))
		if jso, e := rr.Json(); e == nil {
			err := b.Put([]byte(key), []byte(jso))
	
			if err != nil {
				e := fmt.Errorf("Store record failed: %v", rr.Name)
				Log.Debug(e.Error())
				return e
			}
		}

        return nil
    })
 
    return err
}

func DeleteRecord(rr *Record) (err error) {
	checkBdp()
    key, _ := getKey(rr.Name, rr.Type)
    err = bdb.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(rr_bucket))
        err := b.Delete([]byte(key))
 
        if err != nil {
            e := fmt.Errorf("Delete record failed for domain: %v", rr.Name)
            Log.Debug(e.Error())
 
            return e
        }
 
        return nil
    })
 
    return err
}

func createBucket(bucket string) (err error) {
	checkBdp()
    err = bdb.Update(func(tx *bolt.Tx) error {
        _, err := tx.CreateBucketIfNotExists([]byte(bucket))
        if err != nil {
            e := fmt.Errorf("Error creating bucket: %v", bucket)
            Log.Debug(e.Error())
 
            return e
        }
 
        return nil
    })
 
    return err
}

func checkBdp() {
	if bdb == nil {
		err := fmt.Errorf("Internal database not define.")
		panic(err)
	}
}