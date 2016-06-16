// Steven Phillips / elimisteve
// 2016.06.16

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/boltdb/bolt"
)

var (
	minilockBucket = []byte("minilock")

	buckets = [][]byte{
		minilockBucket,
	}

	ErrMinilockIDNotFound = errors.New("miniLock ID not found")
	ErrMinilockIDExists   = errors.New("A miniLock ID already exists for that user")
)

func mustInitDB() *bolt.DB {
	options := &bolt.Options{Timeout: 2 * time.Second}
	dbPath := path.Join(os.Getenv("BOLT_PATH"), "keys.db")

	// Open file
	db, err := bolt.Open(dbPath, 0600, options)
	if err != nil {
		log.Fatalf("Error opening bolt DB: %v", err)
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bkt := range buckets {
			_, err := tx.CreateBucketIfNotExists(bkt)
			if err != nil {
				return fmt.Errorf("Error creating bucket `%s`: %v", bkt, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error creating '%v' boltDB bucket: %v\n", minilockBucket,
			err)
	}

	return db
}

func GetMinilockIDByEmail(db *bolt.DB, email string) (mID string, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(minilockBucket)
		mlockB := bkt.Get([]byte(email))
		if mlockB == nil {
			return ErrMinilockIDNotFound
		}
		mID = string(mlockB)
		return nil
	})
	return mID, err
}

func CreateMinilockIDByEmail(db *bolt.DB, email, mID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(minilockBucket)

		// Only save if user has no existing ID
		oldID := bkt.Get([]byte(email))
		if oldID != nil {
			return ErrMinilockIDExists
		}

		return bkt.Put([]byte(email), []byte(mID))
	})
}
